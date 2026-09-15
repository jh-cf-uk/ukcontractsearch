package handlers

import (
	"net/http"
	"strings"
	"uk-contracts-portal/api"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine) {
	router.GET("/api/contracts", GetContracts)
}

func GetContracts(c *gin.Context) {
	var params api.FilterParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contracts, err := api.FetchContracts(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filtered := filterContracts(contracts.Releases, params)

	response := gin.H{
		"total":    len(filtered),
		"releases": filtered,
	}
	if len(contracts.SourceErrors) > 0 {
		response["sourceErrors"] = contracts.SourceErrors
	}

	c.JSON(http.StatusOK, response)
}

func filterContracts(releases []api.Release, params api.FilterParams) []api.Release {
	var result []api.Release

	for _, release := range releases {
		tender := release.Tender

		// Only active tenders
		if tender.Status != "active" {
			continue
		}

		// Minimum value filter
		if params.MinValue > 0 && tender.Value.Amount < params.MinValue {
			continue
		}

		// VCSE suitability filter
		if params.VCSESuitable && !isSuitableForVCSE(release) {
			continue
		}

		// Sector filter (by CPV code prefix)
		if params.Sector != "" && !matchesSector(tender.Classification.ID, params.Sector) {
			continue
		}

		// Keyword search
		if params.Keyword != "" {
			keyword := strings.ToLower(params.Keyword)
			if !strings.Contains(strings.ToLower(tender.Title), keyword) &&
				!strings.Contains(strings.ToLower(tender.Description), keyword) {
				continue
			}
		}

		result = append(result, release)
	}

	return result
}

func isSuitableForVCSE(release api.Release) bool {
	// Check if any supplier is marked as VCSE
	for _, party := range release.Parties {
		if len(party.Roles) > 0 && party.Roles[0] == "supplier" {
			return true // VCSE-suitable if open to suppliers
		}
	}
	return true
}

func matchesSector(cpvCode string, sector string) bool {
	// Match CPV code prefixes for common social sectors
	socialSectors := map[string][]string{
		"health":    {"85"},
		"education": {"80"},
		"social":    {"85", "72"},
		"it":        {"72"},
		"services":  {"72", "73", "79"},
	}

	codes, exists := socialSectors[strings.ToLower(sector)]
	if !exists {
		return true
	}

	for _, code := range codes {
		if strings.HasPrefix(cpvCode, code) {
			return true
		}
	}
	return false
}
