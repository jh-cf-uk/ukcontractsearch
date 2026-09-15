package api

type ContractsResponse struct {
	URI           string    `json:"uri"`
	Version       string    `json:"version"`
	PublishedDate string    `json:"publishedDate"`
	Publisher     Publisher `json:"publisher"`
	Releases      []Release `json:"releases"`
	Links         Links     `json:"links"`
	SourceErrors  []string  `json:"sourceErrors,omitempty"`
}

type Links struct {
	Next string `json:"next"`
}

type Publisher struct {
	Name   string `json:"name"`
	Scheme string `json:"scheme"`
	UID    string `json:"uid"`
	URI    string `json:"uri"`
}

type Release struct {
	OCID           string   `json:"ocid"`
	ID             string   `json:"id"`
	Language       string   `json:"language"`
	Date           string   `json:"date"`
	Tag            []string `json:"tag"`
	InitiationType string   `json:"initiationType"`
	Tender         Tender   `json:"tender"`
	Buyer          Party    `json:"buyer"`
	Awards         []Award  `json:"awards"`
	Parties        []Party  `json:"parties"`
}

type Tender struct {
	ID                       string         `json:"id"`
	Title                    string         `json:"title"`
	Description              string         `json:"description"`
	Status                   string         `json:"status"`
	DatePublished            string         `json:"datePublished"`
	Classification           Classification `json:"classification"`
	MainProcurementCategory  string         `json:"mainProcurementCategory"`
	ProcurementMethod        string         `json:"procurementMethod"`
	ProcurementMethodDetails string         `json:"procurementMethodDetails"`
	TenderPeriod             Period         `json:"tenderPeriod"`
	ContractPeriod           Period         `json:"contractPeriod"`
	Value                    Value          `json:"value"`
}

type Classification struct {
	Scheme      string `json:"scheme"`
	ID          string `json:"id"`
	Description string `json:"description"`
}

type Period struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type Value struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Party struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Identifier   Identifier   `json:"identifier"`
	Address      Address      `json:"address"`
	ContactPoint ContactPoint `json:"contactPoint"`
	Roles        []string     `json:"roles"`
}

type Identifier struct {
	LegalName string `json:"legalName"`
	Scheme    string `json:"scheme"`
}

type Address struct {
	StreetAddress string `json:"streetAddress"`
	Locality      string `json:"locality"`
	PostalCode    string `json:"postalCode"`
	CountryName   string `json:"countryName"`
}

type ContactPoint struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"telephone"`
}

type Award struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Date      string  `json:"date"`
	Value     Value   `json:"value"`
	Suppliers []Party `json:"suppliers"`
}

type FilterParams struct {
	Keyword      string  `form:"keyword"`
	Sector       string  `form:"sector"`
	MinValue     float64 `form:"minValue"`
	VCSESuitable bool    `form:"vcseSuitable"`
	Page         int     `form:"page"`
}
