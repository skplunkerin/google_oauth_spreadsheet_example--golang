package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Code originally pulled from the following, and then modified for my own
// demonstration:
//   - Google Sheets API - Golang Quickstart:
//     https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample

type Config struct {
	BatchCount          int    `envconfig:"BATCH_COUNT" required:"true" default:"1000"`
	CredentialsFileName string `envconfig:"CREDENTIALS_FILE_NAME" required:"true" default:"credentials.json"`
	// The `SpreadsheetId`/`SheetName` defaults are for a Google Sheets API sample
	// spreadsheet:
	//  - https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	SpreadsheetId string   `envconfig:"SPREADSHEET_ID" required:"true" default:"1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms"`
	SheetName     string   `envconfig:"SHEET_NAME" required:"true" default:"Class Data"`
	Scopes        []string `envconfig:"SCOPES" required:"true" default:"https://www.googleapis.com/auth/drive.readonly"`
}

type Project struct {
	config        Config
	client        *http.Client
	sheetsService *sheets.Service
}

var (
	project Project

	errSheetNotFound = errors.New("sheetTitle not found")
)

// main initializes the project by reading the local `.env`/`credentials.json`
// files and triggering the OAuth authorization if needed; and then prints the
// names and majors of students from a sample spreadsheet.
//
// Edited from original:
// https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample
func main() {
	var c Config
	ctx := context.Background()
	// Load ENV config
	if err := godotenv.Overload(); err != nil {
		// don't care if there is no .env file as we have defaults set.
		if !os.IsNotExist(err) {
			log.Fatalf("Unable to load ENV: %v", err)
		}
	}
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatalf("Unable to get Config: %v", err)
	}
	project.config = c
	b, err := os.ReadFile(project.config.CredentialsFileName)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	fmt.Println("\nThe following scopes will be used:")
	for _, scope := range project.config.Scopes {
		fmt.Println("\t• " + scope)
	}
	fmt.Println()
	// NOTE: if you modify the scopes, delete your previously saved `token.json`
	// file.
	config, err := google.ConfigFromJSON(b, project.config.Scopes...)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	project.client = getClient(config)

	project.sheetsService, err = sheets.NewService(ctx, option.WithHTTPClient(project.client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// Prints the names and majors of students from the sample spreadsheet
	// project.printFromSampleSpreadsheet()
	project.parseFromSampleSpreadsheet()
}

// getSpreadsheetSheetRowCount will return the row count of the `spreadsheetId`
// `sheetTitle` if found; else an `errSheetNotFound` error.
//
// NOTE: the returned row count doesn't account for blank rows; when looping
// through the spreadsheets rows, watch for `len(resp.Values) == 0` to know when
// you're working with a blank row.
func (p Project) getSpreadsheetSheetRowCount() (int, error) {
	rowCount := 0
	resp, err := p.sheetsService.Spreadsheets.Get(p.config.SpreadsheetId).Do()
	if err != nil {
		return 0, err
	}
	// Loop through available sheets, find the `sheetTitle`, and get the
	// `rowCount` if found
	sheetFound := false
	for _, sheet := range resp.Sheets {
		if sheet.Properties.Title == p.config.SheetName {
			sheetFound = true
			rowCount = int(sheet.Properties.GridProperties.RowCount)
			break
		}
	}
	if !sheetFound {
		return 0, errSheetNotFound
	}
	return rowCount, nil
}

// printFromSampleSpreadsheet prints the names and majors of students from the
// Google Sheets API sample spreadsheet:
//  - https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
func (p Project) printFromSampleSpreadsheet() {
	readRange := "Class Data!A2:Z"
	resp, err := p.sheetsService.Spreadsheets.Values.Get("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		fmt.Println("Name, Major:")
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			fmt.Printf("%s, %s\n", row[0], row[4])
		}
	}
}

// ExampleStudent is the structure for the Google API Sample Spreadsheet:
// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
//
// NOTE: parsing to a struct is only possible when we know the Spreadsheet
// structure ahead of time; this wouldn't work if the `spreadsheetId` and
// `sheetTitle` are provided externally.
type ExampleStudent struct {
	StudentName             string
	Gender                  string
	ClassLevel              string
	HomeState               string
	Major                   string
	ExtracurricularActivity string
}

// parseFromSampleSpreadsheet shows how to parse records from a sample
// spreadsheet, using the header (first row) as the keys to map to the
// `ExampleStudent` struct (if found), as well as in a JSON object for when the
// structure isn't known ahead of time.
func (p Project) parseFromSampleSpreadsheet() {
	rowCount, err := p.getSpreadsheetSheetRowCount()
	if err != nil {
		if errors.Is(err, errSheetNotFound) {
			log.Fatalf("Sheet '%s' not found", p.config.SheetName)
		}
		log.Fatalf("Unable to retrieve sheet row count: %v", err)
	}
	sheetHeaders := []interface{}{}
	fmt.Printf("spreadsheetId: %s\n", p.config.SpreadsheetId)
	fmt.Printf("sheetName: %s\n", p.config.SheetName)
	fmt.Printf("rowCount: %d\n", rowCount)
	// Loop through all the rows in batches of `batchCount`
	for i, j := 1, project.config.BatchCount; i <= rowCount; i, j = i+project.config.BatchCount, j+project.config.BatchCount {
		if j >= rowCount {
			j = rowCount
		}
		fmt.Printf("\nfor loop for rows %d-%d\n", i, j)
		// Example result: "'Sheet Name'!A1:Z10"
		readRange := fmt.Sprintf("'%s'!A%d:Z%d", p.config.SheetName, i, j)
		resp, err := p.sheetsService.Spreadsheets.Values.Get(p.config.SpreadsheetId, readRange).Do()
		if err != nil {
			log.Fatalf("Unable to retrieve data from sheet: %v", err)
		}

		// NOTE: this doesn't necessarily mean the end of the sheet has been
		// reached; it's possible there's some blank rows spread throughout the
		// values (as well as blank rows in-between valid rows that also needs to
		// be caught down below while looping through `resp.Values`).
		if len(resp.Values) == 0 {
			fmt.Println("No data found.")
		}
		//
		// Empty rows are removed from Values (if the full batch is empty rows),
		// only loop through them if rows found:
		if len(resp.Values) > 0 {
			for ii, row := range resp.Values {
				// there might be a blank row in-between valid rows, skip to next row if
				// this is blank:
				if len(row) == 0 {
					fmt.Println("Blank row found.")
					continue
				}
				// if this is the first API call, get and print the headers
				if ii == 0 && i == 1 {
					sheetHeaders = row
					// go to next row
					continue
				} else {
					// Parse row to `ExampleStudent` struct:
					//
					// NOTE: parsing to a struct is only possible when we know the
					// Spreadsheet structure ahead of time; this wouldn't work if the
					// `spreadsheetId`/`sheetTitle` were provided externally.
					student := ExampleStudent{}
					// Parse row as JSON object:
					//
					// Parsing as a JSON works great if we don't know the Spreadsheet
					// structure/headers ahead of time, by using the header strings as the
					// keys.
					json := map[string]interface{}{}
					for iii, k := range sheetHeaders {
						// convert key to string:
						var keyString string
						switch key := k.(type) {
						case string:
							keyString = key
						}
						var valueString string
						switch value := row[iii].(type) {
						case string:
							valueString = value
						}
						if keyString != "" && valueString != "" {
							json[keyString] = valueString
						}
						// match the key in order to set the Struct values:
						switch keyString {
						case "Student Name":
							student.StudentName = valueString
						case "Gender":
							student.Gender = valueString
						case "Class Level":
							student.ClassLevel = valueString
						case "Home State":
							student.HomeState = valueString
						case "Major":
							student.Major = valueString
						case "Extracurricular Activity":
							student.ExtracurricularActivity = valueString
						}
					}
					// If the spreadsheet used matches the format of the Google Sheets API
					// sample spreadsheet, print it:
					if student != (ExampleStudent{}) {
						fmt.Printf("ExampleStudent struct:\t%#v\n", student)
					} else {
						fmt.Printf("\t\t json:\t%#v\n\n", json)
					}
				}
			}
		}
	}
	fmt.Printf("\n\nfinished\n\n")
}

// getClient retrieve `token.json` if exists, else triggers `getTokenFromWeb()`
// to save `token.json`, then returns the generated client.
//
// https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample
func getClient(config *oauth2.Config) *http.Client {
	// The file `token.json` stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// getTokenFromWeb request a token from the web, then returns the retrieved
// token.
//
// https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// tokenFromFile retrieves a token from a local file.
//
// https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// saveToken saves a token to a file path.
//
// https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
