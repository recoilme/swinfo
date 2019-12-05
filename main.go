package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"strings"
)

var (
	version = "0.0.1"
	key     = flag.String("key", "", "Similarweb key")
	in      = flag.String("in", "", "path to file with tsv, format: any\tdomain.com\t..")
	out     = flag.String("out", "out.tsv", "path to out.tsv file")
)

//SiteInfo json resp
type SiteInfo struct {
	SiteName        string `json:"site_name"`
	IsSiteVerified  bool   `json:"is_site_verified"`
	Category        string `json:"category"`
	LargeScreenshot string `json:"large_screenshot"`
	ReachMonths     int    `json:"reach_months"`
	DataMonths      int    `json:"data_months"`
	GlobalRank      struct {
		Rank      int `json:"rank"`
		Direction int `json:"direction"`
	} `json:"global_rank"`
	CountryRank struct {
		Country   int `json:"country"`
		Rank      int `json:"rank"`
		Direction int `json:"direction"`
	} `json:"country_rank"`
	CategoryRank struct {
		Category  string `json:"category"`
		Rank      int    `json:"rank"`
		Direction int    `json:"direction"`
	} `json:"category_rank"`
	Title                  string `json:"title"`
	Description            string `json:"description"`
	RedirectURL            string `json:"redirect_url"`
	EstimatedMonthlyVisits struct {
		Two0190501 int `json:"2019-05-01"`
		Two0190601 int `json:"2019-06-01"`
		Two0190701 int `json:"2019-07-01"`
		Two0190801 int `json:"2019-08-01"`
		Two0190901 int `json:"2019-09-01"`
		Two0191001 int `json:"2019-10-01"`
	} `json:"estimated_monthly_visits"`
	Engagments struct {
		Year         int     `json:"year"`
		Month        int     `json:"month"`
		Visits       float64 `json:"visits"`
		TimeOnSite   float64 `json:"time_on_site"`
		PagePerVisit float64 `json:"page_per_visit"`
		BounceRate   float64 `json:"bounce_rate"`
	} `json:"engagments"`
	TopCountryShares []struct {
		Country int     `json:"country"`
		Value   float64 `json:"value"`
		Change  float64 `json:"change"`
	} `json:"top_country_shares"`
	TotalCountries int `json:"total_countries"`
	TrafficSources struct {
		Search        float64 `json:"search"`
		Social        float64 `json:"social"`
		Mail          float64 `json:"mail"`
		PaidReferrals float64 `json:"paid _referrals"`
		Direct        float64 `json:"direct"`
		Referrals     float64 `json:"referrals"`
	} `json:"traffic_sources"`
	ReferralsRatio float64 `json:"referrals_ratio"`
	TopReferring   []struct {
		Site   string  `json:"site"`
		Value  float64 `json:"value"`
		Change float64 `json:"change"`
	} `json:"top_referring"`
	TotalReferring     int           `json:"total_referring"`
	TopDestinations    []interface{} `json:"top_destinations"`
	TotalDestinations  int           `json:"total_destinations"`
	SearchRatio        float64       `json:"search_ratio"`
	TopOrganicKeywords []struct {
		Keyword string  `json:"keyword"`
		Value   float64 `json:"value"`
		Change  float64 `json:"change"`
	} `json:"top_organic_keywords"`
	TopPaidKeywords                   []interface{} `json:"top_paid_keywords"`
	OrganicKeywordsRollingUniqueCount int           `json:"organic_keywords_rolling_unique_count"`
	PaidKeywordsRollingUniqueCount    int           `json:"paid_keywords_rolling_unique_count"`
	OrganicSearchShare                float64       `json:"organic_search_share"`
	PaidSearchShare                   float64       `json:"paid_search_share"`
	SocialRatio                       float64       `json:"social_ratio"`
	TopSocial                         []struct {
		Name   string  `json:"name"`
		Icon   string  `json:"icon"`
		Site   string  `json:"site"`
		Value  float64 `json:"value"`
		Change float64 `json:"change"`
	} `json:"top_social"`
	DisplayAdsRatio               float64       `json:"display_ads_ratio"`
	TopPublishers                 []interface{} `json:"top_publishers"`
	TopAdNetworks                 []interface{} `json:"top_ad_networks"`
	IncomingAdsRollingUniqueCount int           `json:"incoming_ads_rolling_unique_count"`
	AlsoVisitedUniqueCount        int           `json:"also_visited_unique_count"`
	SimilarSites                  []interface{} `json:"similar_sites"`
	SimilarSitesByRank            []interface{} `json:"similar_sites_by_rank"`
	MobileApps                    struct {
	} `json:"mobile_apps"`
	DailyVisitsMinDate string `json:"daily_visits_min_date"`
	DailyVisitsMaxDate string `json:"daily_visits_max_date"`
}

func main() {
	flag.Parse()
	domains, _ := domainstsv(*in)
	sort.Sort(sort.StringSlice(domains))

	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	fmt.Fprintf(buf, "domain\terr\tcategory\tLargeScreenshot\tGlobalRank\tCountryRank\tTitle\tDescription\tVisits\tTimeOnSite\tPagePerVisit\tBounceRate\n")
	for i, d := range domains {
		si, err := info(d)
		if err != nil {
			fmt.Fprintf(buf, "%s\t%s\t\t\t\t\t\t\t\t\t\t\n", d, err)
		}
		rank := 0
		if si.GlobalRank.Rank != 0 {
			rank = si.GlobalRank.Rank
		}
		crank := 0
		if si.CountryRank.Rank != 0 {
			rank = si.CountryRank.Rank
		}
		fmt.Fprintf(buf, "%s\t%s\t%s\t%s\t%d\t%d\t%s\t%s\t%f\t%f\t%f\t%f\n", d, "", si.Category, si.LargeScreenshot, rank, crank,
			si.Title, si.Description, si.Engagments.Visits, si.Engagments.TimeOnSite, si.Engagments.PagePerVisit, si.Engagments.BounceRate)
		if i%100 == 0 {
			buf.Flush()
		}

	}
	buf.Flush()
}
func info(d string) (si SiteInfo, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.similarweb.com/v1/website/%s/general-data/all?api_key=%s", d, *key))
	if err != nil {
		return
	}
	println(d)
	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(contents, &si)
	if err != nil {
		return
	}
	return
}

func domainstsv(path string) (domains []string, err error) {
	f, err := os.Open(*in)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf := bufio.NewReader(f)
	for {
		s, err := buf.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		}
		arr := strings.Split(s, "\t")
		if len(arr) > 1 {
			domains = append(domains, arr[1])
		}
	}

	return
}
