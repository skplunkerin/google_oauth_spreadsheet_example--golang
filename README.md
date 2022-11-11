# Google OAuth and Spreadsheet GoLang Example

This projects shows how to use the Google OAuth process to gain API access for
making a Sheets API request.

- See the following reference examples:
  - [Google Sheets API - Golang Quickstart](https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample)
  - [Google Sheets API - Get spreadsheet values example](https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get#examples)

## Setup the project

1. Clone code, and copy `.env-sample` to `.env.local` and update the ENV values

2. Install modules:

   See [this](https://github.com/golang/go/wiki/Modules) wiki page for
   further information on using Go modules.

   - `go mod download`

     If you have multiple projects with Go that use different versions of the
     same packages in this project it's a good idea to vendor dependencies
     locally.

   - `go mod vendor`

     Anytime you want to make sure your project is running the expected
     dependency versions defined in `go.mod` and `go.sum` run you can verify by
     using this command:

   - `go mod verify`

3. Start the project:

   TODO:

   - ```sh
     TODO:
     ```
