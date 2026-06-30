//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Function struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Tier        string   `json:"tier,omitempty"`
	Price       string   `json:"price,omitempty"`
}

var paidFunctions = []Function{
	// AI/ML Services ($$$ revenue tier)
	{"ai-text-summarize", "AI Text Summarizer", "Summarize long text using AI", "ai", []string{"ai", "summarize", "nlp"}, "premium", "$0.05"},
	{"ai-sentiment-analyzer", "AI Sentiment Analyzer", "Analyze sentiment of text with AI", "ai", []string{"ai", "sentiment", "nlp"}, "premium", "$0.03"},
	{"ai-keyword-extract", "AI Keyword Extractor", "Extract keywords from text using AI", "ai", []string{"ai", "keywords", "nlp"}, "premium", "$0.04"},
	{"ai-language-translate", "AI Language Translator", "Translate text between languages", "ai", []string{"ai", "translate", "nlp"}, "premium", "$0.05"},
	{"ai-topic-classifier", "AI Topic Classifier", "Classify text into topics", "ai", []string{"ai", "classify", "nlp"}, "premium", "$0.04"},
	{"ai-content-generator", "AI Content Generator", "Generate content based on prompts", "ai", []string{"ai", "generate", "content"}, "premium", "$0.08"},
	{"ai-email-writer", "AI Email Writer", "Generate professional emails", "ai", []string{"ai", "email", "write"}, "premium", "$0.06"},
	{"ai-product-description", "AI Product Description", "Generate product descriptions", "ai", []string{"ai", "ecommerce", "product"}, "premium", "$0.05"},
	{"ai-social-media-caption", "AI Social Media Caption", "Create social media captions", "ai", []string{"ai", "social", "caption"}, "premium", "$0.03"},
	{"ai-ad-copy-generator", "AI Ad Copy Generator", "Generate advertising copy", "ai", []string{"ai", "ads", "marketing"}, "premium", "$0.06"},
	{"ai-seo-meta-tags", "AI SEO Meta Tags", "Generate SEO meta tags", "ai", []string{"ai", "seo", "marketing"}, "premium", "$0.03"},
	{"ai-blog-outline", "AI Blog Outline", "Create blog post outlines", "ai", []string{"ai", "blog", "content"}, "premium", "$0.04"},

	// Business/Finance ($ revenue tier)
	{"invoice-generator", "Invoice Generator", "Generate professional invoices as PDF", "business", []string{"invoice", "pdf", "business"}, "pro", "$0.10"},
	{"quote-calculator", "Quote Calculator", "Calculate service quotes", "business", []string{"quote", "calculator", "business"}, "pro", "$0.05"},
	{"profit-margin-calc", "Profit Margin Calculator", "Calculate profit margins", "business", []string{"business", "finance", "calculator"}, "pro", "$0.03"},
	{"break-even-analyzer", "Break Even Analyzer", "Analyze break-even point", "business", []string{"business", "finance"}, "pro", "$0.04"},
	{"roi-calculator", "ROI Calculator", "Calculate return on investment", "business", []string{"business", "finance", "roi"}, "pro", "$0.03"},
	{"loan-payment-calc", "Loan Payment Calculator", "Calculate loan payments", "business", []string{"finance", "loan", "calculator"}, "pro", "$0.03"},
	{"mortgage-calculator", "Mortgage Calculator", "Calculate mortgage payments", "business", []string{"finance", "mortgage"}, "pro", "$0.04"},
	{"currency-converter", "Currency Converter", "Convert between currencies", "finance", []string{"currency", "finance", "exchange"}, "pro", "$0.02"},
	{"tax-calculator", "Tax Calculator", "Calculate taxes", "finance", []string{"tax", "finance", "calculator"}, "pro", "$0.05"},
	{"discount-calculator", "Discount Calculator", "Calculate discounts and sale prices", "business", []string{"calculator", "discount", "business"}, "pro", "$0.02"},
	{"markup-calculator", "Markup Calculator", "Calculate markup percentage", "business", []string{"business", "calculator"}, "pro", "$0.02"},
	{"shipping-cost-calc", "Shipping Cost Calculator", "Calculate shipping costs", "business", []string{"shipping", "calculator", "ecommerce"}, "pro", "$0.03"},

	// Data Processing ($$) revenue tier
	{"csv-to-json", "CSV to JSON Converter", "Convert CSV files to JSON", "data", []string{"csv", "json", "converter"}, "pro", "$0.02"},
	{"json-to-csv", "JSON to CSV Converter", "Convert JSON to CSV", "data", []string{"json", "csv", "converter"}, "pro", "$0.02"},
	{"xml-to-json", "XML to JSON Converter", "Convert XML to JSON", "data", []string{"xml", "json", "converter"}, "pro", "$0.02"},
	{"json-to-xml", "JSON to XML Converter", "Convert JSON to XML", "data", []string{"json", "xml", "converter"}, "pro", "$0.02"},
	{"yaml-to-json", "YAML to JSON Converter", "Convert YAML to JSON", "data", []string{"yaml", "json", "converter"}, "pro", "$0.02"},
	{"json-to-yaml", "JSON to YAML Converter", "Convert JSON to YAML", "data", []string{"json", "yaml", "converter"}, "pro", "$0.02"},
	{"data-validator", "Data Validator", "Validate data against schemas", "data", []string{"validate", "data", "schema"}, "pro", "$0.03"},
	{"data-transformer", "Data Transformer", "Transform data formats", "data", []string{"transform", "data"}, "pro", "$0.03"},
	{"data-aggregator", "Data Aggregator", "Aggregate data from multiple sources", "data", []string{"aggregate", "data"}, "pro", "$0.04"},
	{"data-normalizer", "Data Normalizer", "Normalize data values", "data", []string{"normalize", "data"}, "pro", "$0.03"},

	// E-commerce ($$ revenue tier)
	{"product-pricing-opt", "Product Pricing Optimizer", "Optimize product pricing", "ecommerce", []string{"ecommerce", "pricing", "optimize"}, "pro", "$0.05"},
	{"inventory-turnover", "Inventory Turnover Calculator", "Calculate inventory turnover", "ecommerce", []string{"ecommerce", "inventory"}, "pro", "$0.03"},
	{"cart-abandonment", "Cart Abandonment Analyzer", "Analyze cart abandonment", "ecommerce", []string{"ecommerce", "cart", "analyze"}, "pro", "$0.04"},
	{"product-recommend", "Product Recommender", "Recommend products to customers", "ecommerce", []string{"ecommerce", "recommend", "ai"}, "pro", "$0.06"},
	{"sales-forecast", "Sales Forecast", "Forecast sales trends", "ecommerce", []string{"ecommerce", "forecast", "predict"}, "pro", "$0.05"},

	// Development Tools ($ revenue tier)
	{"code-formatter", "Code Formatter", "Format code in various languages", "dev", []string{"code", "format", "dev"}, "pro", "$0.02"},
	{"regex-tester", "Regex Tester", "Test regular expressions", "dev", []string{"regex", "test", "dev"}, "pro", "$0.01"},
	{"jwt-decode", "JWT Decoder", "Decode JWT tokens", "dev", []string{"jwt", "decode", "auth"}, "pro", "$0.01"},
	{"uuid-generator", "UUID Generator", "Generate unique UUIDs", "dev", []string{"uuid", "generate", "dev"}, "pro", "$0.01"},
	{"hash-generator", "Hash Generator", "Generate cryptographic hashes", "dev", []string{"hash", "crypto", "dev"}, "pro", "$0.01"},
	{"timestamp-convert", "Timestamp Converter", "Convert timestamps", "dev", []string{"timestamp", "convert", "dev"}, "pro", "$0.01"},
	{"base64-encode", "Base64 Encoder", "Encode to base64", "dev", []string{"base64", "encode", "dev"}, "pro", "$0.01"},
	{"base64-decode", "Base64 Decoder", "Decode from base64", "dev", []string{"base64", "decode", "dev"}, "pro", "$0.01"},

	// Analytics & Marketing ($$ revenue tier)
	{"google-ads-estimator", "Google Ads Estimator", "Estimate Google Ads costs and clicks", "marketing", []string{"ads", "google", "estimate"}, "pro", "$0.03"},
	{"seo-analyzer", "SEO Analyzer", "Analyze SEO metrics for URLs", "marketing", []string{"seo", "analyze", "marketing"}, "pro", "$0.04"},
	{"keyword-difficulty", "Keyword Difficulty Score", "Calculate keyword difficulty scores", "marketing", []string{"seo", "keyword", "marketing"}, "pro", "$0.02"},
	{"backlink-checker", "Backlink Checker", "Check backlinks for domains", "marketing", []string{"seo", "backlink", "marketing"}, "pro", "$0.03"},
	{"competitor-analyzer", "Competitor Analyzer", "Analyze competitor websites", "marketing", []string{"seo", "competitor", "marketing"}, "pro", "$0.05"},
	{"keyword-rank-tracker", "Keyword Rank Tracker", "Track keyword rankings", "marketing", []string{"seo", "rank", "marketing"}, "pro", "$0.03"},
	{"content-gap-analyzer", "Content Gap Analyzer", "Find content gaps vs competitors", "marketing", []string{"seo", "content", "marketing"}, "pro", "$0.04"},
	{"email-deliverability", "Email Deliverability Checker", "Check email deliverability", "marketing", []string{"email", "deliverability", "marketing"}, "pro", "$0.02"},
	{"social-media-scheduler", "Social Media Scheduler", "Schedule social media posts", "marketing", []string{"social", "scheduler", "marketing"}, "pro", "$0.03"},
	{"landing-page-analyzer", "Landing Page Analyzer", "Analyze landing page conversion", "marketing", []string{"landing", "analyze", "marketing"}, "pro", "$0.04"},

	// Legal & Compliance ($$ revenue tier)
	{"privacy-policy-gen", "Privacy Policy Generator", "Generate privacy policy documents", "legal", []string{"legal", "privacy", "document"}, "pro", "$0.08"},
	{"terms-of-service", "Terms of Service Generator", "Generate terms of service", "legal", []string{"legal", "terms", "document"}, "pro", "$0.08"},
	{"cookie-consent-gen", "Cookie Consent Banner", "Generate cookie consent banners", "legal", []string{"legal", "cookie", "gdpr"}, "pro", "$0.05"},
	{"nda-generator", "NDA Generator", "Generate non-disclosure agreements", "legal", []string{"legal", "nda", "document"}, "pro", "$0.07"},
	{"contract-analyzer", "Contract Analyzer", "Analyze legal contracts", "legal", []string{"legal", "contract", "analyze"}, "pro", "$0.10"},
	{"compliance-checker", "Compliance Checker", "Check regulatory compliance", "legal", []string{"legal", "compliance", "gdpr"}, "pro", "$0.06"},

	// Healthcare ($$ revenue tier)
	{"bmi-calculator", "BMI Calculator", "Calculate body mass index", "health", []string{"health", "bmi", "calculator"}, "pro", "$0.01"},
	{"calorie-estimator", "Calorie Estimator", "Estimate daily calorie needs", "health", []string{"health", "calorie", "fitness"}, "pro", "$0.02"},
	{"water-intake-calc", "Water Intake Calculator", "Calculate daily water needs", "health", []string{"health", "water", "fitness"}, "pro", "$0.01"},
	{"sleep-quality-analyzer", "Sleep Quality Analyzer", "Analyze sleep quality", "health", []string{"health", "sleep", "analyze"}, "pro", "$0.02"},
	{"heart-rate-zone", "Heart Rate Zone Calculator", "Calculate heart rate zones", "health", []string{"health", "fitness", "heart"}, "pro", "$0.01"},
	{"macro-nutrient", "Macro Nutrient Calculator", "Calculate macro nutrient needs", "health", []string{"health", "fitness", "nutrition"}, "pro", "$0.02"},

	// Education ($$ revenue tier)
	{"flashcard-generator", "Flashcard Generator", "Generate flashcards from text", "education", []string{"education", "flashcard", "study"}, "pro", "$0.02"},
	{"quiz-generator", "Quiz Generator", "Generate quizzes from content", "education", []string{"education", "quiz", "test"}, "pro", "$0.03"},
	{"study-schedule", "Study Schedule Planner", "Create study schedules", "education", []string{"education", "schedule", "planner"}, "pro", "$0.02"},
	{"citation-generator", "Citation Generator", "Generate academic citations", "education", []string{"education", "citation", "academic"}, "pro", "$0.02"},
	{"plagiarism-checker", "Plagiarism Checker", "Check for plagiarism", "education", []string{"education", "plagiarism", "check"}, "pro", "$0.04"},

	// Real Estate ($$ revenue tier)
	{"mortgage-qualifier", "Mortgage Qualifier", "Qualify for mortgage loans", "realestate", []string{"realestate", "mortgage", "loan"}, "pro", "$0.05"},
	{"rent-estimator", "Rent Estimator", "Estimate rental property values", "realestate", []string{"realestate", "rent", "estimate"}, "pro", "$0.04"},
	{"property-tax-calc", "Property Tax Calculator", "Calculate property taxes", "realestate", []string{"realestate", "tax", "calculator"}, "pro", "$0.03"},
	{"home-affordability", "Home Affordability Calculator", "Calculate affordable home prices", "realestate", []string{"realestate", "afford", "calculator"}, "pro", "$0.04"},
	{"roi-property-calc", "Property ROI Calculator", "Calculate real estate ROI", "realestate", []string{"realestate", "roi", "calculator"}, "pro", "$0.03"},

	// Additional 25 functions to reach 100
	{"email-template-gen", "Email Template Generator", "Generate email templates", "marketing", []string{"email", "template", "marketing"}, "pro", "$0.03"},
	{"linkedin-post-gen", "LinkedIn Post Generator", "Generate LinkedIn posts", "social", []string{"social", "linkedin", "content"}, "pro", "$0.02"},
	{"twitter-thread-gen", "Twitter Thread Generator", "Generate Twitter threads", "social", []string{"social", "twitter", "content"}, "pro", "$0.02"},
	{"instagram-caption-gen", "Instagram Caption Generator", "Generate Instagram captions", "social", []string{"social", "instagram", "content"}, "pro", "$0.02"},
	{"facebook-post-gen", "Facebook Post Generator", "Generate Facebook posts", "social", []string{"social", "facebook", "content"}, "pro", "$0.02"},
	{"video-script-gen", "Video Script Generator", "Generate video scripts", "content", []string{"video", "script", "content"}, "pro", "$0.04"},
	{"podcast-outline-gen", "Podcast Outline Generator", "Generate podcast outlines", "content", []string{"podcast", "outline", "content"}, "pro", "$0.03"},
	{"ebook-outline-gen", "Ebook Outline Generator", "Generate ebook outlines", "content", []string{"ebook", "outline", "content"}, "pro", "$0.04"},
	{"whitepaper-gen", "Whitepaper Generator", "Generate whitepaper content", "content", []string{"whitepaper", "content", "business"}, "pro", "$0.08"},
	{"press-release-gen", "Press Release Generator", "Generate press releases", "marketing", []string{"press", "release", "business"}, "pro", "$0.06"},
	{"business-name-gen", "Business Name Generator", "Generate business names", "business", []string{"business", "name", "generator"}, "pro", "$0.02"},
	{"domain-name-gen", "Domain Name Generator", "Generate domain name ideas", "business", []string{"domain", "name", "generator"}, "pro", "$0.02"},
	{"logo-concept-gen", "Logo Concept Generator", "Generate logo concepts", "design", []string{"logo", "design", "generator"}, "pro", "$0.05"},
	{"color-palette-gen", "Color Palette Generator", "Generate color palettes", "design", []string{"color", "palette", "design"}, "pro", "$0.02"},
	{"resume-builder", "Resume Builder", "Build professional resumes", "career", []string{"resume", "career", "document"}, "pro", "$0.06"},
	{"cover-letter-gen", "Cover Letter Generator", "Generate cover letters", "career", []string{"cover", "letter", "career"}, "pro", "$0.05"},
	{"interview-questions", "Interview Question Generator", "Generate interview questions", "career", []string{"interview", "questions", "career"}, "pro", "$0.03"},
	{"job-description-gen", "Job Description Generator", "Generate job descriptions", "career", []string{"job", "description", "hr"}, "pro", "$0.04"},
	{"payroll-calculator", "Payroll Calculator", "Calculate payroll costs", "business", []string{"payroll", "calculator", "business"}, "pro", "$0.03"},
	{"benefits-calculator", "Benefits Calculator", "Calculate employee benefits", "business", []string{"benefits", "calculator", "hr"}, "pro", "$0.03"},
	{"employee-scheduler", "Employee Scheduler", "Create employee schedules", "business", []string{"schedule", "employee", "hr"}, "pro", "$0.04"},
	{"performance-review", "Performance Review Generator", "Generate performance reviews", "business", []string{"review", "performance", "hr"}, "pro", "$0.05"},
	{"expense-report-gen", "Expense Report Generator", "Generate expense reports", "business", []string{"expense", "report", "finance"}, "pro", "$0.03"},
	{"invoice-template", "Invoice Template Generator", "Generate invoice templates", "business", []string{"invoice", "template", "business"}, "pro", "$0.04"},
	{"proposal-generator", "Proposal Generator", "Generate business proposals", "business", []string{"proposal", "business", "document"}, "pro", "$0.05"},
	{"contract-generator", "Contract Generator", "Generate contracts", "legal", []string{"contract", "legal", "document"}, "pro", "$0.07"},
	{"affidavit-generator", "Affidavit Generator", "Generate affidavits", "legal", []string{"affidavit", "legal", "document"}, "pro", "$0.06"},
}

func parsePrice(priceStr string) float64 {
	// Remove $ prefix and parse as float
	priceStr = strings.TrimPrefix(priceStr, "$")
	val, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0
	}
	return val
}

func main() {
	baseDir := "/home/micro/projects/functionfly/functions/functionfly"

	for _, fn := range paidFunctions {
		dir := filepath.Join(baseDir, fn.Name)
		os.MkdirAll(dir, 0755)

		pricePerCall := parsePrice(fn.Price)

		// Create main.py
		mainPy := fmt.Sprintf(`import json

def handler(event):
    if isinstance(event, dict):
        data = event.get("data", "")
    else:
        data = ""

    # TODO: Implement %s logic
    result = {"ok": True, "result": data, "tier": "%s"}
    return result
`, fn.Title, fn.Tier)
		os.WriteFile(filepath.Join(dir, "main.py"), []byte(mainPy), 0644)

		// Create functionfly.jsonc with price_per_call as numeric
		manifest := map[string]interface{}{
			"author":         "functionfly",
			"name":           fn.Name,
			"version":        "1.0.0",
			"runtime":        "python3.12",
			"title":          fn.Title,
			"description":    fn.Description,
			"category":       fn.Category,
			"tags":           fn.Tags,
			"tier":           fn.Tier,
			"price_per_call": pricePerCall,
			"input": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"data": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"output": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"ok": map[string]interface{}{
						"type": "boolean",
					},
					"result": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"deterministic": true,
			"idempotent":    true,
			"cache_ttl":     0,
			"timeout_ms":    5000,
			"memory_mb":     128,
		}

		manifestJSON, _ := json.MarshalIndent(manifest, "", "  ")
		os.WriteFile(filepath.Join(dir, "functionfly.jsonc"), []byte(string(manifestJSON)+"\n"), 0644)

		fmt.Printf("Created: %s\n", fn.Name)
	}

	fmt.Printf("\nCreated %d paid functions\n", len(paidFunctions))
}
