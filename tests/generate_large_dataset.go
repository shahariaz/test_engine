package main

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

var (
	// Sample data pools for realistic generation
	firstNames = []string{
		"Aamir", "Shabana", "Nasir", "Mehreen", "Tariq", "Fatima", "Hassan", "Ayesha",
		"Rahman", "Zainab", "Ibrahim", "Sadia", "Omar", "Rukhsana", "Ali", "Nadia",
		"Ahmed", "Samina", "Khalid", "Farah", "Imran", "Saira", "Yasir", "Hina",
		"Rashid", "Uzma", "Salman", "Fariha", "Kamran", "Shazia", "Bilal", "Rubina",
		"Shahid", "Bushra", "Naveed", "Kiran", "Asif", "Lubna", "Adnan", "Rafia",
		"Jane", "John", "Sarah", "Michael", "Emma", "David", "Lisa", "James",
		"Maria", "Robert", "Jennifer", "William", "Linda", "Richard", "Patricia", "Thomas",
		"Rakib", "Shahin", "Nasreen", "Karim", "Rashida", "Selim", "Ruma", "Hanif",
		"Sultana", "Mizanur", "Rokeya", "Abdur", "Shaheen", "Rezaul", "Hosneara", "Mofizul",
	}
	
	lastNames = []string{
		"Khan", "Ahmed", "Ali", "Rahman", "Hassan", "Sheikh", "Chowdhury", "Islam",
		"Begum", "Khatun", "Uddin", "Alam", "Haque", "Sultana", "Molla", "Sarkar",
		"Das", "Roy", "Pal", "Ghosh", "Mukherjee", "Chakraborty", "Sen", "Gupta",
		"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis",
		"Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas",
	}
	
	countries = []string{
		"Bangladesh", "India", "Pakistan", "United States", "United Kingdom", 
		"Canada", "Australia", "Germany", "France", "Japan", "Singapore", "Malaysia",
	}
	
	cities = map[string][]string{
		"Bangladesh": {"Dhaka", "Chittagong", "Sylhet", "Rajshahi", "Khulna", "Barisal", "Rangpur", "Mymensingh"},
		"India": {"Mumbai", "Delhi", "Bangalore", "Kolkata", "Chennai", "Hyderabad", "Pune", "Ahmedabad"},
		"Pakistan": {"Karachi", "Lahore", "Islamabad", "Rawalpindi", "Faisalabad", "Multan", "Peshawar", "Quetta"},
		"United States": {"New York", "Los Angeles", "Chicago", "Houston", "Phoenix", "Philadelphia", "San Antonio", "San Diego"},
		"United Kingdom": {"London", "Birmingham", "Manchester", "Glasgow", "Liverpool", "Leeds", "Sheffield", "Edinburgh"},
		"Canada": {"Toronto", "Montreal", "Vancouver", "Calgary", "Edmonton", "Ottawa", "Winnipeg", "Quebec City"},
		"Australia": {"Sydney", "Melbourne", "Brisbane", "Perth", "Adelaide", "Gold Coast", "Newcastle", "Canberra"},
		"Germany": {"Berlin", "Hamburg", "Munich", "Cologne", "Frankfurt", "Stuttgart", "Düsseldorf", "Leipzig"},
		"France": {"Paris", "Marseille", "Lyon", "Toulouse", "Nice", "Nantes", "Montpellier", "Strasbourg"},
		"Japan": {"Tokyo", "Osaka", "Kyoto", "Yokohama", "Kobe", "Nagoya", "Sapporo", "Fukuoka"},
		"Singapore": {"Singapore"},
		"Malaysia": {"Kuala Lumpur", "George Town", "Ipoh", "Shah Alam", "Petaling Jaya", "Klang", "Johor Bahru", "Seberang Perai"},
	}
	
	deviceModels = []string{
		"iPhone 15 Pro", "iPhone 15", "iPhone 14 Pro", "iPhone 14", "iPhone 13 Pro", "iPhone 13", "iPhone 12",
		"Samsung Galaxy S24", "Samsung Galaxy S23", "Samsung Galaxy S22", "Samsung Galaxy A54", "Samsung Galaxy A34",
		"Google Pixel 8", "Google Pixel 7", "OnePlus 11", "OnePlus 10", "Xiaomi 13", "Xiaomi 12",
		"Huawei P60", "Huawei Mate 60", "OPPO Find X6", "Vivo X90", "Realme GT3", "Nothing Phone 2",
	}
	
	appVersions = []string{
		"6.2.1", "6.2.0", "6.1.5", "6.1.4", "6.1.3", "6.1.2", "6.1.1", "6.1.0",
		"6.0.9", "6.0.8", "6.0.7", "6.0.6", "6.0.5", "6.0.4", "6.0.3", "6.0.2", "6.0.1", "6.0.0",
		"5.9.8", "5.9.7", "5.9.6", "5.9.5", "5.8.2", "5.8.1", "5.8.0", "5.7.5", "5.7.4", "5.7.3",
	}
	
	packages = []string{"Premium", "Basic", "Student", "Family"}
	subscriptionStatuses = []string{"active", "trial", "expired", "cancelled", "paused"}
	
	contentTitles = []string{
		"আমার সোনার বাংলা", "কাবুলিওয়ালা", "পদ্মা নদীর মাঝি", "শেষের কবিতা", "চোখের বালি",
		"গোরা", "ঘরে বাইরে", "নৌকাডুবি", "যোগাযোগ", "মালঞ্চ", "রক্তকরবী", "ডাকঘর",
		"Tech Talks with Experts", "Digital Bangladesh", "Startup Stories", "AI Revolution",
		"Coding Bootcamp", "Data Science Masterclass", "Blockchain Explained", "Cloud Computing",
		"Dhallywood Classic", "Bengali Romance", "Historical Drama", "Comedy Nights",
		"Action Thriller", "Family Drama", "Musical Journey", "Documentary Series",
		"Crime Investigation", "Medical Drama", "Legal Thriller", "Sci-Fi Adventure",
	}
	
	genres = []string{
		"Drama", "Comedy", "Action", "Romance", "Thriller", "Documentary", 
		"Educational", "Musical", "Historical", "Sci-Fi", "Crime", "Family",
	}
	
	contentTypes = []string{"movie", "series", "documentary", "educational"}
)

func randomChoice(slice []string) string {
	return slice[rand.Intn(len(slice))]
}

func randomInt(min, max int) int {
	return rand.Intn(max-min+1) + min
}

func randomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func generateUID(prefix string, id int) string {
	return fmt.Sprintf("_:%s_%d", prefix, id)
}

func generateCustomers(count int) []string {
	var customers []string
	
	for i := 1; i <= count; i++ {
		uid := generateUID("customer", i)
		name := fmt.Sprintf("%s %s", randomChoice(firstNames), randomChoice(lastNames))
		email := strings.ToLower(fmt.Sprintf("%s.%s%d@example.com", 
			randomChoice(firstNames), randomChoice(lastNames), randomInt(1, 999)))
		age := randomInt(13, 65)
		country := randomChoice(countries)
		city := randomChoice(cities[country])
		
		customer := fmt.Sprintf(`%s <chorki_customers.name> "%s" .
%s <chorki_customers.email> "%s" .
%s <chorki_customers.age> "%d"^^<xs:int> .
%s <chorki_customers.country> "%s" .
%s <chorki_customers.city> "%s" .
%s <dgraph.type> "chorki_customers" .`,
			uid, name, uid, email, uid, age, uid, country, uid, city, uid)
		
		customers = append(customers, customer)
	}
	
	return customers
}

func generateSubscriptions(customerCount int) []string {
	var subscriptions []string
	
	// Generate subscriptions for 80% of customers
	subscriptionCount := int(float64(customerCount) * 0.8)
	
	for i := 1; i <= subscriptionCount; i++ {
		uid := generateUID("subscription", i)
		customerUID := generateUID("customer", randomInt(1, customerCount))
		
		pkg := randomChoice(packages)
		status := randomChoice(subscriptionStatuses)
		
		// Adjust probabilities for more realistic data
		if rand.Float64() < 0.6 {
			status = "active"
		} else if rand.Float64() < 0.3 {
			status = "trial"
		}
		
		price := 0.0
		switch pkg {
		case "Premium":
			price = randomFloat(299, 599)
		case "Basic":
			price = randomFloat(99, 199)
		case "Student":
			price = randomFloat(49, 99)
		case "Family":
			price = randomFloat(399, 799)
		}
		
		subscription := fmt.Sprintf(`%s <chorki_subscriptions.customer> %s .
%s <chorki_subscriptions.subscribed_package> "%s" .
%s <chorki_subscriptions.subscription_status> "%s" .
%s <chorki_subscriptions.price> "%.2f"^^<xs:float> .
%s <dgraph.type> "chorki_subscriptions" .
%s <~chorki_customers.subscriptions> %s .`,
			uid, customerUID, uid, pkg, uid, status, uid, price, uid, uid, customerUID)
		
		subscriptions = append(subscriptions, subscription)
	}
	
	return subscriptions
}

func generateDevices(customerCount int) []string {
	var devices []string
	
	// Generate 1-3 devices per customer
	deviceID := 1
	for customerID := 1; customerID <= customerCount; customerID++ {
		numDevices := randomInt(1, 3)
		customerUID := generateUID("customer", customerID)
		
		for j := 0; j < numDevices; j++ {
			uid := generateUID("device", deviceID)
			model := randomChoice(deviceModels)
			
			var osType string
			if strings.Contains(model, "iPhone") {
				osType = "iOS"
			} else {
				osType = "Android"
			}
			
			appVersion := randomChoice(appVersions)
			
			device := fmt.Sprintf(`%s <chorki_devices.customer> %s .
%s <chorki_devices.device> "%s" .
%s <chorki_devices.model> "%s" .
%s <chorki_devices.app_version> "%s" .
%s <dgraph.type> "chorki_devices" .
%s <~chorki_customers.devices> %s .`,
				uid, customerUID, uid, osType, uid, model, uid, appVersion, uid, uid, customerUID)
			
			devices = append(devices, device)
			deviceID++
		}
	}
	
	return devices
}

func generateContent(count int) []string {
	var contents []string
	
	for i := 1; i <= count; i++ {
		uid := generateUID("content", i)
		title := randomChoice(contentTitles)
		contentType := randomChoice(contentTypes)
		genre := randomChoice(genres)
		duration := randomInt(15, 180) // 15 minutes to 3 hours
		rating := randomFloat(3.0, 9.5)
		year := randomInt(2018, 2024)
		
		content := fmt.Sprintf(`%s <chorki_contents.title> "%s" .
%s <chorki_contents.type> "%s" .
%s <chorki_contents.genre> "%s" .
%s <chorki_contents.duration_minutes> "%d"^^<xs:int> .
%s <chorki_contents.rating> "%.1f"^^<xs:float> .
%s <chorki_contents.release_year> "%d"^^<xs:int> .
%s <dgraph.type> "chorki_contents" .`,
			uid, title, uid, contentType, uid, genre, uid, duration, uid, rating, uid, year, uid)
		
		contents = append(contents, content)
	}
	
	return contents
}

func generateWatchHistories(customerCount, contentCount int) []string {
	var watchHistories []string
	
	// Generate watch histories for customers
	historyID := 1
	for customerID := 1; customerID <= customerCount; customerID++ {
		// 70% of customers have watch history
		if rand.Float64() < 0.7 {
			numWatches := randomInt(1, 10)
			customerUID := generateUID("customer", customerID)
			
			for j := 0; j < numWatches; j++ {
				uid := generateUID("watch_history", historyID)
				contentUID := generateUID("content", randomInt(1, contentCount))
				
				watchedMinutes := randomInt(5, 120)
				watchedDate := fmt.Sprintf("2024-%02d-%02d", randomInt(1, 12), randomInt(1, 28))
				
				watchHistory := fmt.Sprintf(`%s <chorki_watch_histories.customer> %s .
%s <chorki_watch_histories.content> %s .
%s <chorki_watch_histories.watched_minutes> "%d"^^<xs:int> .
%s <chorki_watch_histories.watched_date> "%s"^^<xs:dateTime> .
%s <dgraph.type> "chorki_watch_histories" .
%s <~chorki_customers.watch_histories> %s .`,
					uid, customerUID, uid, contentUID, uid, watchedMinutes, uid, watchedDate, uid, uid, customerUID)
				
				watchHistories = append(watchHistories, watchHistory)
				historyID++
			}
		}
	}
	
	return watchHistories
}

func main() {
	rand.Seed(time.Now().UnixNano())
	
	fmt.Println("🚀 Generating large dataset with 10,000+ records...")
	
	// Generate different entity counts
	customerCount := 10000
	contentCount := 500
	
	fmt.Printf("📊 Generating %d customers...\n", customerCount)
	customers := generateCustomers(customerCount)
	
	fmt.Printf("💳 Generating subscriptions...\n")
	subscriptions := generateSubscriptions(customerCount)
	
	fmt.Printf("📱 Generating devices...\n")
	devices := generateDevices(customerCount)
	
	fmt.Printf("🎬 Generating %d content items...\n", contentCount)
	contents := generateContent(contentCount)
	
	fmt.Printf("👀 Generating watch histories...\n")
	watchHistories := generateWatchHistories(customerCount, contentCount)
	
	// Write to file
	file, err := os.Create("dgraph/large_sample_data.rdf")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	
	fmt.Println("💾 Writing data to file...")
	
	// Write all customers
	for _, customer := range customers {
		file.WriteString(customer + "\n\n")
	}
	
	// Write all subscriptions
	for _, subscription := range subscriptions {
		file.WriteString(subscription + "\n\n")
	}
	
	// Write all devices
	for _, device := range devices {
		file.WriteString(device + "\n\n")
	}
	
	// Write all contents
	for _, content := range contents {
		file.WriteString(content + "\n\n")
	}
	
	// Write all watch histories
	for _, watchHistory := range watchHistories {
		file.WriteString(watchHistory + "\n\n")
	}
	
	fmt.Printf("✅ Dataset generated successfully!\n")
	fmt.Printf("📈 Statistics:\n")
	fmt.Printf("   - Customers: %d\n", len(customers))
	fmt.Printf("   - Subscriptions: %d\n", len(subscriptions))
	fmt.Printf("   - Devices: %d\n", len(devices))
	fmt.Printf("   - Content: %d\n", len(contents))
	fmt.Printf("   - Watch Histories: %d\n", len(watchHistories))
	fmt.Printf("   - Total Records: %d\n", len(customers)+len(subscriptions)+len(devices)+len(contents)+len(watchHistories))
	fmt.Printf("💾 Data saved to: dgraph/large_sample_data.rdf\n")
}