package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const ContractsFinderAPI = "https://www.contractsfinder.service.gov.uk/Published/Notices/OCDS/Search"
const FindATenderAPI = "https://www.find-tender.service.gov.uk/api/1.0/ocdsReleasePackages"

var httpClient = &http.Client{Timeout: 30 * time.Second}
var contractsCache struct {
	sync.Mutex
	response  *ContractsResponse
	fetchedAt time.Time
}

const contractsCacheDuration = 5 * time.Minute

func FetchContracts(params FilterParams) (*ContractsResponse, error) {
	contractsCache.Lock()
	if contractsCache.response != nil && time.Since(contractsCache.fetchedAt) < contractsCacheDuration {
		response := contractsCache.response
		contractsCache.Unlock()
		return response, nil
	}
	contractsCache.Unlock()

	publishedFrom := time.Now().AddDate(0, -6, 0).Format("2006-01-02")
	publishedTo := time.Now().Format("2006-01-02")

	contractsFinder, contractsFinderErr := fetchContractsFinder(publishedFrom, publishedTo)
	findATender, findATenderErr := fetchFindATender(publishedFrom, publishedTo)

	if contractsFinderErr != nil && findATenderErr != nil {
		return nil, fmt.Errorf("Contracts Finder: %v; Find a Tender: %v", contractsFinderErr, findATenderErr)
	}

	var releases []Release
	var sourceErrors []string
	if contractsFinderErr != nil {
		sourceErrors = append(sourceErrors, contractsFinderErr.Error())
	} else {
		releases = append(releases, contractsFinder.Releases...)
	}
	if findATenderErr != nil {
		sourceErrors = append(sourceErrors, findATenderErr.Error())
	} else {
		releases = append(releases, findATender.Releases...)
	}

	response := &ContractsResponse{
		Releases:     releases,
		SourceErrors: sourceErrors,
	}

	contractsCache.Lock()
	contractsCache.response = response
	contractsCache.fetchedAt = time.Now()
	contractsCache.Unlock()

	return response, nil
}

func fetchContractsFinder(publishedFrom string, publishedTo string) (*ContractsResponse, error) {
	var allReleases []Release
	cursor := ""

	for page := 0; page < 10; page++ {
		query := url.Values{}
		query.Set("publishedFrom", publishedFrom)
		query.Set("publishedTo", publishedTo)
		query.Set("limit", "100")

		if cursor != "" {
			query.Set("cursor", cursor)
		}

		fullURL := ContractsFinderAPI + "?" + query.Encode()
		resp, err := httpClient.Get(fullURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch Contracts Finder: %w", err)
		}

		body, err := readResponseBody(resp, "Contracts Finder")
		if err != nil {
			return nil, err
		}

		var response ContractsResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Contracts Finder JSON: %w", err)
		}

		allReleases = append(allReleases, response.Releases...)
		if response.Links.Next == "" {
			break
		}

		nextURL, err := url.Parse(response.Links.Next)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Contracts Finder next link: %w", err)
		}
		cursor = nextURL.Query().Get("cursor")
		if cursor == "" {
			break
		}
	}

	return &ContractsResponse{Releases: allReleases}, nil
}

func fetchFindATender(updatedFrom string, updatedTo string) (*ContractsResponse, error) {
	var allReleases []Release
	start, err := time.Parse("2006-01-02", updatedFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid Find a Tender start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", updatedTo)
	if err != nil {
		return nil, fmt.Errorf("invalid Find a Tender end date: %w", err)
	}

	for !start.After(end) {
		chunkEnd := start.AddDate(0, 0, 6)
		if chunkEnd.After(end) {
			chunkEnd = end
		}

		cursor := ""
		for page := 0; page < 10; page++ {
			query := url.Values{}
			query.Set("updatedFrom", start.Format("2006-01-02")+"T00:00:00")
			query.Set("updatedTo", chunkEnd.Format("2006-01-02")+"T23:59:59")
			query.Set("limit", "100")
			query.Set("stages", "tender")
			if cursor != "" {
				query.Set("cursor", cursor)
			}

			resp, err := httpClient.Get(FindATenderAPI + "?" + query.Encode())
			if err != nil {
				return nil, fmt.Errorf("failed to fetch Find a Tender: %w", err)
			}

			body, err := readResponseBody(resp, "Find a Tender")
			if err != nil {
				return nil, err
			}

			var response ContractsResponse
			if err := json.Unmarshal(body, &response); err != nil {
				return nil, fmt.Errorf("failed to parse Find a Tender JSON: %w", err)
			}

			allReleases = append(allReleases, response.Releases...)
			if response.Links.Next == "" {
				break
			}

			nextURL, err := url.Parse(response.Links.Next)
			if err != nil {
				return nil, fmt.Errorf("failed to parse Find a Tender next link: %w", err)
			}
			cursor = nextURL.Query().Get("cursor")
			if cursor == "" {
				break
			}
		}

		start = chunkEnd.AddDate(0, 0, 1)
	}

	return &ContractsResponse{Releases: allReleases}, nil
}

func readResponseBody(resp *http.Response, source string) ([]byte, error) {
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s returned HTTP %s: %s", source, resp.Status, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", source, err)
	}
	return body, nil
}
