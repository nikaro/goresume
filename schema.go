package main

type Resume struct {
	Schema       string        `json:"$schema,omitempty"`
	Awards       []Award       `json:"awards,omitempty"`
	Basics       Basics        `json:"basics,omitempty"`
	Certificates []Certificate `json:"certificates,omitempty"`
	Education    []Education   `json:"education,omitempty"`
	Interests    []Interest    `json:"interests,omitempty"`
	Languages    []Language    `json:"languages,omitempty"`
	Meta         Meta          `json:"meta,omitempty"`
	Projects     []Project     `json:"projects,omitempty"`
	Publications []Publication `json:"publications,omitempty"`
	References   []Reference   `json:"references,omitempty"`
	Skills       []Skill       `json:"skills,omitempty"`
	Volunteer    []Volunteer   `json:"volunteer,omitempty"`
	Work         []Work        `json:"work,omitempty"`
}

type Award struct {
	Title   string `json:"title,omitempty"`
	Date    string `json:"date,omitempty"`
	Awarder string `json:"awarder,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type Basics struct {
	Name     string    `json:"name,omitempty"`
	Label    string    `json:"label,omitempty"`
	Image    string    `json:"image,omitempty"`
	Email    string    `json:"email,omitempty"`
	Phone    string    `json:"phone,omitempty"`
	Url      string    `json:"url,omitempty"`
	Summary  string    `json:"summary,omitempty"`
	Location Location  `json:"location,omitempty"`
	Profiles []Profile `json:"profiles,omitempty"`
}

type Location struct {
	Address     string `json:"address,omitempty"`
	PostalCode  string `json:"postalCode,omitempty"`
	City        string `json:"city,omitempty"`
	CountryCode string `json:"countryCode,omitempty"`
	Region      string `json:"region,omitempty"`
}

type Profile struct {
	Network  string `json:"network,omitempty"`
	Url      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
}

type Certificate struct {
	Name   string `json:"name,omitempty"`
	Date   string `json:"date,omitempty"`
	Url    string `json:"url,omitempty"`
	Issuer string `json:"issuer,omitempty"`
}

type Education struct {
	Institution string   `json:"institution,omitempty"`
	Url         string   `json:"url,omitempty"`
	Area        string   `json:"area,omitempty"`
	StudyType   string   `json:"studyType,omitempty"`
	StartDate   string   `json:"startDate,omitempty"`
	EndDate     string   `json:"endDate,omitempty"`
	Score       string   `json:"score,omitempty"`
	Courses     []string `json:"courses,omitempty"`
}

type Interest struct {
	Name     string   `json:"interest,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}
type Language struct {
	Language string `json:"language,omitempty"`
	Fluency  string `json:"fluency,omitempty"`
}

type Meta struct {
	Canonical    string `json:"canonical,omitempty"`
	Version      string `json:"version,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
}

type Project struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Highlights  []string `json:"highlights,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
	StartDate   string   `json:"startDate,omitempty"`
	EndDate     string   `json:"endDate,omitempty"`
	Url         string   `json:"url,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Entity      string   `json:"entity,omitempty"`
	Type        string   `json:"type,omitempty"`
}

type Publication struct {
	Name        string `json:"name,omitempty"`
	Publisher   string `json:"publisher,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	Url         string `json:"url,omitempty"`
	Summary     string `json:"summary,omitempty"`
}

type Reference struct {
	Name      string `json:"name,omitempty"`
	Reference string `json:"reference,omitempty"`
}

type Skill struct {
	Name     string   `json:"name,omitempty"`
	Level    string   `json:"level,omitempty"`
	Keywords []string `json:"keywords,omitempty"`
}

type Volunteer struct {
	Organization string   `json:"organization,omitempty"`
	Position     string   `json:"position,omitempty"`
	Url          string   `json:"url,omitempty"`
	StartDate    string   `json:"startDate,omitempty"`
	EndDate      string   `json:"endDate,omitempty"`
	Summary      string   `json:"summary,omitempty"`
	Highlights   []string `json:"highlights,omitempty"`
}

type Work struct {
	Name        string   `json:"name,omitempty"`
	Location    string   `json:"location,omitempty"`
	Description string   `json:"description,omitempty"`
	Position    string   `json:"position,omitempty"`
	Url         string   `json:"url,omitempty"`
	StartDate   string   `json:"startDate,omitempty"`
	EndDate     string   `json:"endDate,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Highlights  []string `json:"highlights,omitempty"`
	Keywords    []string `json:"keywords,omitempty"`
}
