package pkg

import (
	"fmt"
	"math/rand"
	"time"
)

// DataProvider contains all the reference data for generation
type DataProvider struct {
	Countries      []string
	Cities         map[string][]string
	Devices        []string
	DeviceModels   map[string][]string
	AppVersions    []string
	Packages       []string
	Statuses       []string
	PaymentMethods []string
	ContentTypes   []string
	Genres         []string
	Languages      []string
	Qualities      []string
	FirstNames     []string
	LastNames      []string
}

// NewDataProvider creates a new data provider with predefined data
func NewDataProvider() *DataProvider {
	return &DataProvider{
		Countries: []string{
			"Bangladesh", "USA", "Canada", "UK", "India", "Australia",
			"Germany", "France", "Japan", "Brazil", "Mexico", "Italy",
			"Spain", "Netherlands", "Sweden", "Norway", "Denmark", "Finland",
		},
		Cities: map[string][]string{
			"Bangladesh": {"Dhaka", "Chittagong", "Sylhet", "Rajshahi", "Khulna", "Barisal", "Rangpur", "Mymensingh"},
			"USA":        {"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio"},
			"Canada":     {"Toronto", "Montreal", "Vancouver", "Calgary", "Ottawa", "Edmonton", "Mississauga"},
			"UK":         {"London", "Birmingham", "Manchester", "Glasgow", "Newcastle", "Liverpool", "Bristol"},
			"India":      {"Mumbai", "Delhi", "Bangalore", "Hyderabad", "Chennai", "Kolkata", "Pune"},
		},
		Devices: []string{"iOS", "Android", "Web", "SmartTV", "Roku", "FireStick", "AppleTV", "PlayStation", "Xbox"},
		DeviceModels: map[string][]string{
			"iOS":     {"iPhone 15", "iPhone 14", "iPhone 13", "iPhone 12", "iPad Pro", "iPad Air", "iPad Mini"},
			"Android": {"Samsung Galaxy S24", "Google Pixel 8", "OnePlus 12", "Xiaomi 14", "Samsung Galaxy Tab", "Huawei P60"},
			"Web":     {"Chrome", "Firefox", "Safari", "Edge", "Opera"},
			"SmartTV": {"Samsung Smart TV", "LG Smart TV", "Sony Smart TV", "TCL Smart TV", "Hisense Smart TV"},
		},
		AppVersions: []string{"5.0.0", "5.1.0", "5.2.0", "5.3.0", "5.4.0", "6.0.0", "6.1.0", "6.2.0", "7.0.0"},
		Packages:    []string{"Free", "Basic", "Premium", "VIP", "Family", "Student", "Monthly", "Yearly"},
		Statuses:    []string{"active", "inactive", "trial", "expired", "suspended", "cancelled", "pending"},
		PaymentMethods: []string{
			"credit_card", "debit_card", "paypal", "bank_transfer",
			"mobile_payment", "crypto", "apple_pay", "google_pay", "stripe",
		},
		ContentTypes: []string{
			"movie", "series", "documentary", "short_film", "music_video",
			"live_stream", "podcast", "anime", "web_series", "reality_show",
		},
		Genres: []string{
			"Action", "Drama", "Comedy", "Horror", "Thriller", "Romance",
			"Sci-Fi", "Fantasy", "Adventure", "Documentary", "Animation",
			"Crime", "Mystery", "Family", "Musical", "Biography", "War", "Western",
		},
		Languages: []string{
			"English", "Bengali", "Hindi", "Spanish", "French", "German",
			"Japanese", "Korean", "Arabic", "Portuguese", "Italian", "Russian",
		},
		Qualities: []string{"360p", "480p", "720p", "1080p", "4K", "8K"},
		FirstNames: []string{
			"John", "Jane", "Michael", "Sarah", "David", "Emily", "Chris", "Jessica",
			"Robert", "Ashley", "William", "Amanda", "James", "Stephanie", "Daniel", "Jennifer",
			"Matthew", "Lisa", "Anthony", "Karen", "Mark", "Nancy", "Steven", "Betty",
			"Rihad", "Tanvir", "Sakib", "Rashid", "Fahim", "Naeem", "Shahin", "Rafi",
			"Sabbir", "Masum", "Fatima", "Ayesha", "Rahul", "Priya", "Ahmed", "Nadia",
		},
		LastNames: []string{
			"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
			"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson",
			"Rahman", "Ahmed", "Khan", "Hossain", "Islam", "Alam", "Ali", "Hassan",
			"Sheikh", "Uddin", "Chowdhury", "Sharma", "Gupta", "Singh", "Kumar",
		},
	}
}

// generateSingleCustomer generates a single customer with all relationships
func (g *Generator) generateSingleCustomer(id int) map[string]interface{} {
	data := NewDataProvider()

	country := g.randomChoice(data.Countries)
	cityList := data.Cities[country]
	if cityList == nil {
		cityList = []string{"Unknown City"}
	}

	device := g.randomChoice(data.Devices)
	firstName := g.randomChoice(data.FirstNames)
	lastName := g.randomChoice(data.LastNames)
	name := fmt.Sprintf("%s %s", firstName, lastName)
	email := fmt.Sprintf("%s.%s%d@example.com", firstName, lastName, rand.Intn(1000))

	customer := map[string]interface{}{
		"dgraph.type":                      []interface{}{"chorki_customers"},
		"chorki_customers.id":              fmt.Sprintf("CUST%05d", id),
		"chorki_customers.name":            name,
		"chorki_customers.email":           email,
		"chorki_customers.age":             rand.Intn(65) + 13, // Age between 13 and 77
		"chorki_customers.country":         country,
		"chorki_customers.device":          device,
		"chorki_customers.app_version":     g.randomChoice(data.AppVersions),
		"chorki_customers.last_login_days": rand.Intn(90),
		"chorki_customers.is_active":       rand.Float64() < g.config.Distribution.ActiveCustomerRatio,
		"chorki_customers.created_at":      g.randomDate(3),
		"chorki_customers.updated_at":      g.randomDate(1),
	}

	// Generate subscriptions (based on distribution config)
	numSubs := g.randomBetween(
		g.config.Distribution.MinSubscriptionsPerCustomer,
		g.config.Distribution.MaxSubscriptionsPerCustomer,
	)

	var subscriptions []map[string]interface{}
	for j := 0; j < numSubs; j++ {
		subscription := map[string]interface{}{
			"dgraph.type":                         []interface{}{"chorki_subscriptions"},
			"chorki_subscriptions.id":             fmt.Sprintf("SUB%05d_%d", id, j+1),
			"chorki_subscriptions.package":        g.randomChoice(data.Packages),
			"chorki_subscriptions.status":         g.randomChoice(data.Statuses),
			"chorki_subscriptions.start_date":     g.randomDate(2),
			"chorki_subscriptions.end_date":       g.randomDate(1),
			"chorki_subscriptions.price":          float64(rand.Intn(50)+5) + rand.Float64(),
			"chorki_subscriptions.currency":       "USD",
			"chorki_subscriptions.payment_method": g.randomChoice(data.PaymentMethods),
			"chorki_subscriptions.auto_renewal":   rand.Float32() < 0.6,
			"chorki_subscriptions.trial_period":   rand.Float32() < 0.3,
			"chorki_subscriptions.created_at":     g.randomDate(2),
			"chorki_subscriptions.updated_at":     g.randomDate(1),
		}
		subscriptions = append(subscriptions, subscription)
	}
	customer["chorki_customers.subscriptions"] = subscriptions

	// Generate watch histories
	numWatch := g.randomBetween(
		g.config.Distribution.MinWatchHistoriesPerCustomer,
		g.config.Distribution.MaxWatchHistoriesPerCustomer,
	)

	var watchHistories []map[string]interface{}
	for j := 0; j < numWatch; j++ {
		duration := rand.Intn(180) + 30 // 30-210 minutes
		watched := rand.Intn(duration)

		watchHistory := map[string]interface{}{
			"dgraph.type":                                  []interface{}{"chorki_watch_histories"},
			"chorki_watch_histories.id":                    fmt.Sprintf("WATCH%05d_%d", id, j+1),
			"chorki_watch_histories.content_id":            fmt.Sprintf("CONTENT%d", rand.Intn(g.config.Contents)+1),
			"chorki_watch_histories.content_title":         fmt.Sprintf("Content Title %d", rand.Intn(500)+1),
			"chorki_watch_histories.type":                  g.randomChoice(data.ContentTypes),
			"chorki_watch_histories.genre":                 g.randomChoices(data.Genres, rand.Intn(3)+1),
			"chorki_watch_histories.watch_duration":        watched,
			"chorki_watch_histories.completion_percentage": float64(watched) / float64(duration) * 100,
			"chorki_watch_histories.watch_date":            g.randomDate(1),
			"chorki_watch_histories.created_at":            g.randomDate(1),
		}
		watchHistories = append(watchHistories, watchHistory)
	}
	customer["chorki_customers.watch_histories"] = watchHistories

	// Generate devices
	numDevices := g.randomBetween(
		g.config.Distribution.MinDevicesPerCustomer,
		g.config.Distribution.MaxDevicesPerCustomer,
	)

	var devices []map[string]interface{}
	for j := 0; j < numDevices; j++ {
		deviceType := g.randomChoice(data.Devices)
		deviceModelList := data.DeviceModels[deviceType]
		if deviceModelList == nil {
			deviceModelList = []string{"Unknown Model"}
		}

		device := map[string]interface{}{
			"dgraph.type":                 []interface{}{"chorki_devices"},
			"chorki_devices.id":           fmt.Sprintf("DEV%05d_%d", id, j+1),
			"chorki_devices.device_type":  deviceType,
			"chorki_devices.device_model": g.randomChoice(deviceModelList),
			"chorki_devices.os_version":   fmt.Sprintf("%d.%d.%d", rand.Intn(5)+10, rand.Intn(10), rand.Intn(10)),
			"chorki_devices.app_version":  g.randomChoice(data.AppVersions),
			"chorki_devices.is_active":    rand.Float32() < 0.9,
			"chorki_devices.last_seen":    g.randomDate(1),
			"chorki_devices.created_at":   g.randomDate(2),
			"chorki_devices.updated_at":   g.randomDate(1),
		}
		devices = append(devices, device)
	}
	customer["chorki_customers.devices"] = devices

	return customer
}

// generateSingleContent generates a single content item
func (g *Generator) generateSingleContent(id int) map[string]interface{} {
	data := NewDataProvider()

	contentType := g.randomChoice(data.ContentTypes)
	contentGenres := g.randomChoices(data.Genres, rand.Intn(3)+1)

	content := map[string]interface{}{
		"dgraph.type":                   []interface{}{"chorki_contents"},
		"chorki_contents.id":            fmt.Sprintf("CONTENT%d", id),
		"chorki_contents.title":         fmt.Sprintf("%s Title %d", contentType, id),
		"chorki_contents.type":          contentType,
		"chorki_contents.genre":         contentGenres,
		"chorki_contents.duration":      rand.Intn(180) + 30,              // 30-210 minutes
		"chorki_contents.rating":        float64(rand.Intn(50)+50) / 10.0, // 5.0-10.0
		"chorki_contents.release_date":  g.randomDate(365 * 5),            // Within last 5 years
		"chorki_contents.language":      g.randomChoice(data.Languages),
		"chorki_contents.description":   fmt.Sprintf("Description for %s content %d", contentType, id),
		"chorki_contents.thumbnail_url": fmt.Sprintf("https://example.com/thumbnails/%d.jpg", id),
		"chorki_contents.video_url":     fmt.Sprintf("https://example.com/videos/%d.mp4", id),
		"chorki_contents.is_premium":    rand.Float64() < g.config.Distribution.PremiumContentRatio,
		"chorki_contents.created_at":    g.randomDate(365 * 3),
		"chorki_contents.updated_at":    g.randomDate(365),
	}

	return content
}

// Helper methods

func (g *Generator) randomChoice(slice []string) string {
	if len(slice) == 0 {
		return ""
	}
	return slice[rand.Intn(len(slice))]
}

func (g *Generator) randomChoices(slice []string, count int) []string {
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

func (g *Generator) randomBetween(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min+1) + min
}

func (g *Generator) randomDate(daysBack int) string {
	now := time.Now()
	minTime := now.AddDate(0, 0, -daysBack)
	delta := now.Unix() - minTime.Unix()
	sec := rand.Int63n(delta) + minTime.Unix()
	return time.Unix(sec, 0).Format(time.RFC3339)
}
