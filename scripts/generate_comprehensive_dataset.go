package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

// Customer represents a customer entity
type Customer struct {
	UID              string         `json:"uid,omitempty"`
	DType            []string       `json:"dgraph.type"`
	ID               string         `json:"chorki_customers.id"`
	Name             string         `json:"chorki_customers.name"`
	Email            string         `json:"chorki_customers.email"`
	Phone            string         `json:"chorki_customers.phone"`
	Age              int            `json:"chorki_customers.age"`
	Country          string         `json:"chorki_customers.country"`
	City             string         `json:"chorki_customers.city"`
	Device           string         `json:"chorki_customers.device"`
	AppVersion       string         `json:"chorki_customers.app_version"`
	RegistrationDate string         `json:"chorki_customers.registration_date"`
	LastLogin        string         `json:"chorki_customers.last_login"`
	LastLoginDays    int            `json:"chorki_customers.last_login_days"`
	IsActive         bool           `json:"chorki_customers.is_active"`
	CreatedAt        string         `json:"chorki_customers.created_at"`
	UpdatedAt        string         `json:"chorki_customers.updated_at"`
	Subscriptions    []Subscription `json:"chorki_customers.subscriptions,omitempty"`
	WatchHistories   []WatchHistory `json:"chorki_customers.watch_histories,omitempty"`
	Devices          []Device       `json:"chorki_customers.devices,omitempty"`
}

// Subscription represents a subscription entity
type Subscription struct {
	UID           string   `json:"uid,omitempty"`
	DType         []string `json:"dgraph.type"`
	ID            string   `json:"chorki_subscriptions.id"`
	CustomerID    string   `json:"chorki_subscriptions.customer_id"`
	Package       string   `json:"chorki_subscriptions.package"`
	Status        string   `json:"chorki_subscriptions.status"`
	StartDate     string   `json:"chorki_subscriptions.start_date"`
	EndDate       string   `json:"chorki_subscriptions.end_date"`
	Price         float64  `json:"chorki_subscriptions.price"`
	Currency      string   `json:"chorki_subscriptions.currency"`
	PaymentMethod string   `json:"chorki_subscriptions.payment_method"`
	AutoRenewal   bool     `json:"chorki_subscriptions.auto_renewal"`
	TrialPeriod   bool     `json:"chorki_subscriptions.trial_period"`
	CreatedAt     string   `json:"chorki_subscriptions.created_at"`
	UpdatedAt     string   `json:"chorki_subscriptions.updated_at"`
}

// WatchHistory represents a watch history entity
type WatchHistory struct {
	UID                  string   `json:"uid,omitempty"`
	DType                []string `json:"dgraph.type"`
	ID                   string   `json:"chorki_watch_histories.id"`
	CustomerID           string   `json:"chorki_watch_histories.customer_id"`
	ContentID            int      `json:"chorki_watch_histories.content_id"`
	ContentTitle         string   `json:"chorki_watch_histories.content_title"`
	Type                 string   `json:"chorki_watch_histories.type"`
	Genre                []string `json:"chorki_watch_histories.genre"`
	DurationWatched      int      `json:"chorki_watch_histories.duration_watched"`
	TotalDuration        int      `json:"chorki_watch_histories.total_duration"`
	CompletionPercentage float64  `json:"chorki_watch_histories.completion_percentage"`
	WatchDate            string   `json:"chorki_watch_histories.watch_date"`
	DeviceUsed           string   `json:"chorki_watch_histories.device_used"`
	Quality              string   `json:"chorki_watch_histories.quality"`
	CreatedAt            string   `json:"chorki_watch_histories.created_at"`
}

// Content represents a content entity
type Content struct {
	UID          string   `json:"uid,omitempty"`
	DType        []string `json:"dgraph.type"`
	ID           int      `json:"chorki_contents.id"`
	Title        string   `json:"chorki_contents.title"`
	Type         string   `json:"chorki_contents.type"`
	Genre        []string `json:"chorki_contents.genre"`
	Duration     int      `json:"chorki_contents.duration"`
	ReleaseYear  int      `json:"chorki_contents.release_year"`
	Rating       float64  `json:"chorki_contents.rating"`
	Language     string   `json:"chorki_contents.language"`
	Country      string   `json:"chorki_contents.country"`
	Director     string   `json:"chorki_contents.director"`
	Cast         []string `json:"chorki_contents.cast"`
	Description  string   `json:"chorki_contents.description"`
	ThumbnailURL string   `json:"chorki_contents.thumbnail_url"`
	VideoURL     string   `json:"chorki_contents.video_url"`
	IsPremium    bool     `json:"chorki_contents.is_premium"`
	IsActive     bool     `json:"chorki_contents.is_active"`
	CreatedAt    string   `json:"chorki_contents.created_at"`
	UpdatedAt    string   `json:"chorki_contents.updated_at"`
}

// Device represents a device entity
type Device struct {
	UID         string   `json:"uid,omitempty"`
	DType       []string `json:"dgraph.type"`
	ID          string   `json:"chorki_devices.id"`
	CustomerID  string   `json:"chorki_devices.customer_id"`
	DeviceType  string   `json:"chorki_devices.device_type"`
	DeviceModel string   `json:"chorki_devices.device_model"`
	OSVersion   string   `json:"chorki_devices.os_version"`
	AppVersion  string   `json:"chorki_devices.app_version"`
	PushToken   string   `json:"chorki_devices.push_token"`
	IsActive    bool     `json:"chorki_devices.is_active"`
	LastUsed    string   `json:"chorki_devices.last_used"`
	CreatedAt   string   `json:"chorki_devices.created_at"`
	UpdatedAt   string   `json:"chorki_devices.updated_at"`
}

var (
	countries = []string{"Bangladesh", "USA", "Canada", "UK", "India", "Australia", "Germany", "France", "Japan", "Brazil", "Mexico", "Italy", "Spain", "Netherlands", "Sweden"}
	cities    = map[string][]string{
		"Bangladesh": {"Dhaka", "Chittagong", "Sylhet", "Rajshahi", "Khulna", "Barisal", "Rangpur"},
		"USA":        {"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio"},
		"Canada":     {"Toronto", "Montreal", "Vancouver", "Calgary", "Ottawa", "Edmonton", "Mississauga"},
		"UK":         {"London", "Birmingham", "Manchester", "Glasgow", "Newcastle", "Liverpool", "Bristol"},
		"India":      {"Mumbai", "Delhi", "Bangalore", "Hyderabad", "Chennai", "Kolkata", "Pune"},
	}
	devices      = []string{"iOS", "Android", "Web", "SmartTV", "Roku", "FireStick", "AppleTV"}
	deviceModels = map[string][]string{
		"iOS":     {"iPhone 14", "iPhone 13", "iPhone 12", "iPad Pro", "iPad Air", "iPad Mini"},
		"Android": {"Samsung Galaxy S23", "Google Pixel 7", "OnePlus 11", "Xiaomi 13", "Samsung Galaxy Tab"},
		"Web":     {"Chrome", "Firefox", "Safari", "Edge"},
		"SmartTV": {"Samsung Smart TV", "LG Smart TV", "Sony Smart TV", "TCL Smart TV"},
	}
	appVersions    = []string{"5.0.0", "5.1.0", "5.2.0", "5.3.0", "5.4.0", "6.0.0", "6.1.0"}
	packages       = []string{"Free", "Basic", "Premium", "VIP", "Family", "Student"}
	statuses       = []string{"active", "inactive", "trial", "expired", "suspended", "cancelled"}
	paymentMethods = []string{"credit_card", "debit_card", "paypal", "bank_transfer", "mobile_payment", "crypto"}
	contentTypes   = []string{"movie", "series", "documentary", "short_film", "music_video", "live_stream", "podcast"}
	genres         = []string{"Action", "Drama", "Comedy", "Horror", "Thriller", "Romance", "Sci-Fi", "Fantasy", "Adventure", "Documentary", "Animation", "Crime", "Mystery", "Family", "Musical"}
	languages      = []string{"English", "Bengali", "Hindi", "Spanish", "French", "German", "Japanese", "Korean", "Arabic", "Portuguese"}
	qualities      = []string{"480p", "720p", "1080p", "4K", "8K"}
	firstNames     = []string{"John", "Jane", "Michael", "Sarah", "David", "Emily", "Chris", "Jessica", "Robert", "Ashley", "William", "Amanda", "James", "Stephanie", "Daniel", "Jennifer", "Matthew", "Lisa", "Anthony", "Karen", "Mark", "Nancy", "Steven", "Betty", "Paul", "Helen", "Andrew", "Sandra", "Joshua", "Donna", "Kenneth", "Carol", "Kevin", "Ruth", "Brian", "Sharon", "George", "Michelle", "Edward", "Laura", "Ronald", "Sarah", "Timothy", "Kimberly", "Jason", "Deborah", "Jeffrey", "Dorothy", "Ryan", "Lisa", "Jacob", "Nancy", "Gary", "Karen", "Nicholas", "Betty", "Eric", "Helen", "Jonathan", "Sandra", "Stephen", "Donna", "Larry", "Carol", "Justin", "Ruth", "Scott", "Sharon", "Brandon", "Michelle", "Benjamin", "Laura", "Samuel", "Sarah", "Gregory", "Kimberly", "Alexander", "Deborah", "Patrick", "Dorothy", "Frank", "Lisa", "Raymond", "Nancy", "Jack", "Karen", "Dennis", "Betty", "Jerry", "Helen", "Tyler", "Sandra", "Aaron", "Donna", "Jose", "Carol", "Henry", "Ruth", "Adam", "Sharon", "Douglas", "Michelle", "Nathan", "Laura", "Peter", "Sarah", "Zachary", "Kimberly", "Kyle", "Deborah", "Noah", "Dorothy", "Alan", "Lisa", "Ethan", "Nancy", "Jeremy", "Karen", "Rihad", "Tanvir", "Sakib", "Rashid", "Fahim", "Naeem", "Shahin", "Rafi", "Sabbir", "Masum"}
	lastNames      = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson", "Walker", "Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores", "Green", "Adams", "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell", "Carter", "Roberts", "Rahman", "Ahmed", "Khan", "Hossain", "Islam", "Alam", "Ali", "Hassan", "Sheikh", "Uddin", "Chowdhury"}
)

func randomChoice(slice []string) string {
	return slice[rand.Intn(len(slice))]
}

func randomChoices(slice []string, count int) []string {
	if count >= len(slice) {
		return slice
	}

	result := make([]string, 0, count)
	used := make(map[int]bool)

	for len(result) < count {
		idx := rand.Intn(len(slice))
		if !used[idx] {
			used[idx] = true
			result = append(result, slice[idx])
		}
	}

	return result
}

func randomDate(yearsBack int) string {
	now := time.Now()
	minTime := now.AddDate(-yearsBack, 0, 0)
	delta := now.Unix() - minTime.Unix()
	sec := rand.Int63n(delta) + minTime.Unix()
	return time.Unix(sec, 0).Format(time.RFC3339)
}

func generateCustomers(count int) []Customer {
	customers := make([]Customer, count)

	for i := 0; i < count; i++ {
		country := randomChoice(countries)
		cityList := cities[country]
		if cityList == nil {
			cityList = []string{"Unknown City"}
		}

		device := randomChoice(devices)
		deviceModelList := deviceModels[device]
		if deviceModelList == nil {
			deviceModelList = []string{"Unknown Model"}
		}

		firstName := randomChoice(firstNames)
		lastName := randomChoice(lastNames)
		name := fmt.Sprintf("%s %s", firstName, lastName)
		email := fmt.Sprintf("%s.%s%d@example.com", firstName, lastName, rand.Intn(1000))

		customers[i] = Customer{
			DType:            []string{"chorki_customers"},
			ID:               fmt.Sprintf("CUST%05d", i+1),
			Name:             name,
			Email:            email,
			Phone:            fmt.Sprintf("+%d%d%d%d%d%d%d%d%d", rand.Intn(9)+1, rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10), rand.Intn(10)),
			Age:              rand.Intn(65) + 13, // Age between 13 and 77
			Country:          country,
			City:             randomChoice(cityList),
			Device:           device,
			AppVersion:       randomChoice(appVersions),
			RegistrationDate: randomDate(3),
			LastLogin:        randomDate(1),
			LastLoginDays:    rand.Intn(90),
			IsActive:         rand.Float32() < 0.8, // 80% active
			CreatedAt:        randomDate(3),
			UpdatedAt:        randomDate(1),
		}

		// Generate subscriptions (1-3 per customer)
		numSubs := rand.Intn(3) + 1
		for j := 0; j < numSubs; j++ {
			subscription := Subscription{
				DType:         []string{"chorki_subscriptions"},
				ID:            fmt.Sprintf("SUB%05d", i*10+j+1),
				CustomerID:    customers[i].ID,
				Package:       randomChoice(packages),
				Status:        randomChoice(statuses),
				StartDate:     randomDate(2),
				EndDate:       randomDate(1),
				Price:         float64(rand.Intn(50)+5) + rand.Float64(), // $5-$55
				Currency:      "USD",
				PaymentMethod: randomChoice(paymentMethods),
				AutoRenewal:   rand.Float32() < 0.6,
				TrialPeriod:   rand.Float32() < 0.3,
				CreatedAt:     randomDate(2),
				UpdatedAt:     randomDate(1),
			}
			customers[i].Subscriptions = append(customers[i].Subscriptions, subscription)
		}

		// Generate watch histories (2-10 per customer)
		numWatch := rand.Intn(9) + 2
		for j := 0; j < numWatch; j++ {
			duration := rand.Intn(180) + 30 // 30-210 minutes
			watched := rand.Intn(duration)

			watchHistory := WatchHistory{
				DType:                []string{"chorki_watch_histories"},
				ID:                   fmt.Sprintf("WATCH%05d", i*10+j+1),
				CustomerID:           customers[i].ID,
				ContentID:            rand.Intn(500) + 1,
				ContentTitle:         fmt.Sprintf("Content Title %d", rand.Intn(500)+1),
				Type:                 randomChoice(contentTypes),
				Genre:                randomChoices(genres, rand.Intn(3)+1),
				DurationWatched:      watched,
				TotalDuration:        duration,
				CompletionPercentage: float64(watched) / float64(duration) * 100,
				WatchDate:            randomDate(1),
				DeviceUsed:           customers[i].Device,
				Quality:              randomChoice(qualities),
				CreatedAt:            randomDate(1),
			}
			customers[i].WatchHistories = append(customers[i].WatchHistories, watchHistory)
		}

		// Generate devices (1-3 per customer)
		numDevices := rand.Intn(3) + 1
		for j := 0; j < numDevices; j++ {
			deviceType := randomChoice(devices)
			deviceModelList := deviceModels[deviceType]
			if deviceModelList == nil {
				deviceModelList = []string{"Unknown Model"}
			}

			device := Device{
				DType:       []string{"chorki_devices"},
				ID:          fmt.Sprintf("DEV%05d", i*10+j+1),
				CustomerID:  customers[i].ID,
				DeviceType:  deviceType,
				DeviceModel: randomChoice(deviceModelList),
				OSVersion:   fmt.Sprintf("%d.%d.%d", rand.Intn(5)+10, rand.Intn(10), rand.Intn(10)),
				AppVersion:  randomChoice(appVersions),
				PushToken:   fmt.Sprintf("token_%d_%d_%d", i, j, rand.Intn(10000)),
				IsActive:    rand.Float32() < 0.9,
				LastUsed:    randomDate(1),
				CreatedAt:   randomDate(2),
				UpdatedAt:   randomDate(1),
			}
			customers[i].Devices = append(customers[i].Devices, device)
		}
	}

	return customers
}

func generateContents(count int) []Content {
	contents := make([]Content, count)

	for i := 0; i < count; i++ {
		contentType := randomChoice(contentTypes)
		contentGenres := randomChoices(genres, rand.Intn(3)+1)

		contents[i] = Content{
			DType:        []string{"chorki_contents"},
			ID:           i + 1,
			Title:        fmt.Sprintf("%s Title %d", contentType, i+1),
			Type:         contentType,
			Genre:        contentGenres,
			Duration:     rand.Intn(180) + 30,              // 30-210 minutes
			ReleaseYear:  rand.Intn(20) + 2005,             // 2005-2024
			Rating:       float64(rand.Intn(50)+50) / 10.0, // 5.0-10.0
			Language:     randomChoice(languages),
			Country:      randomChoice(countries),
			Director:     fmt.Sprintf("Director %d", rand.Intn(100)+1),
			Cast:         randomChoices(firstNames, rand.Intn(5)+2),
			Description:  fmt.Sprintf("Description for %s content %d", contentType, i+1),
			ThumbnailURL: fmt.Sprintf("https://example.com/thumbnails/%d.jpg", i+1),
			VideoURL:     fmt.Sprintf("https://example.com/videos/%d.mp4", i+1),
			IsPremium:    rand.Float32() < 0.4,  // 40% premium
			IsActive:     rand.Float32() < 0.95, // 95% active
			CreatedAt:    randomDate(3),
			UpdatedAt:    randomDate(1),
		}
	}

	return contents
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("🚀 Generating comprehensive dataset with 100+ customers and 500+ content items...")

	// Generate data
	customers := generateCustomers(120)
	contents := generateContents(500)

	// Combine all data
	allData := make([]interface{}, 0)

	// Add customers with their nested relationships
	for _, customer := range customers {
		allData = append(allData, customer)
	}

	// Add standalone contents
	for _, content := range contents {
		allData = append(allData, content)
	}

	// Write to JSON file
	jsonData, err := json.MarshalIndent(allData, "", "  ")
	if err != nil {
		fmt.Printf("❌ Error marshaling JSON: %v\n", err)
		return
	}

	err = os.WriteFile("comprehensive_dataset.json", jsonData, 0644)
	if err != nil {
		fmt.Printf("❌ Error writing file: %v\n", err)
		return
	}

	fmt.Printf("✅ Generated comprehensive dataset:\n")
	fmt.Printf("   📊 Customers: %d\n", len(customers))
	fmt.Printf("   📊 Contents: %d\n", len(contents))

	totalSubs := 0
	totalWatch := 0
	totalDevices := 0

	for _, customer := range customers {
		totalSubs += len(customer.Subscriptions)
		totalWatch += len(customer.WatchHistories)
		totalDevices += len(customer.Devices)
	}

	fmt.Printf("   📊 Subscriptions: %d\n", totalSubs)
	fmt.Printf("   📊 Watch Histories: %d\n", totalWatch)
	fmt.Printf("   📊 Devices: %d\n", totalDevices)
	fmt.Printf("   📊 Total Records: %d\n", len(customers)+len(contents)+totalSubs+totalWatch+totalDevices)
	fmt.Printf("\n📁 File saved as: comprehensive_dataset.json\n")

	// Generate sample queries for testing
	sampleQueries := []map[string]interface{}{
		{
			"name":        "Young Premium Users",
			"description": "Find users under 30 with Premium subscriptions",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "age", "op": "<", "value": 30},
							{"field": "subscribed_package", "op": "=", "value": "Premium"},
						},
					},
				},
			},
		},
		{
			"name":        "Multi-Device Users",
			"description": "Users with iOS or Android devices and high app version",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "device", "op": "=", "value": "iOS"},
							{"field": "device", "op": "=", "value": "Android"},
						},
					},
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "app_version", "op": ">=", "value": "6.0.0"},
						},
					},
				},
			},
		},
		{
			"name":        "Active Movie Watchers",
			"description": "Active users who watch movies",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "content_type", "op": "=", "value": "movie"},
							{"field": "subscription_status", "op": "=", "value": "active"},
						},
					},
				},
			},
		},
		{
			"name":        "Complex Nested Query",
			"description": "Young users from specific countries with premium subscriptions or trial status",
			"query": map[string]interface{}{
				"combine_with": "AND",
				"groups": []map[string]interface{}{
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "country", "op": "IN", "value": []string{"USA", "Canada", "UK"}},
						},
					},
					{
						"combine_with": "AND",
						"filters": []map[string]interface{}{
							{"field": "age", "op": ">=", "value": 18},
							{"field": "age", "op": "<=", "value": 35},
						},
					},
					{
						"combine_with": "OR",
						"filters": []map[string]interface{}{
							{"field": "subscribed_package", "op": "=", "value": "Premium"},
							{"field": "subscription_status", "op": "=", "value": "trial"},
						},
					},
				},
			},
		},
	}

	queriesJson, _ := json.MarshalIndent(sampleQueries, "", "  ")
	err = os.WriteFile("sample_queries.json", queriesJson, 0644)
	if err == nil {
		fmt.Printf("📁 Sample queries saved as: sample_queries.json\n")
	}

	fmt.Println("\n🎯 Ready to load data into Dgraph!")
	fmt.Println("   1. Update schema: curl -X POST localhost:8080/admin/schema --data-binary '@dgraph/schema.graphql'")
	fmt.Println("   2. Load data: curl -X POST localhost:8080/mutate?commitNow=true -H 'Content-Type: application/json' --data-binary '@comprehensive_dataset.json'")
}
