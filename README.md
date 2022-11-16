# Google OAuth and Spreadsheet GoLang Example

This projects shows how to use the Google OAuth process to gain API access for
making a Sheets API request.

- See the following reference examples:
  - [Google Sheets API - Golang Quickstart](https://developers.google.com/sheets/api/quickstart/go#step_3_set_up_the_sample)
  - [Google Sheets API - Get spreadsheet values example](https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get#examples)

## Setup the project

1. Clone code and setup local files:

   - Copy `.env-sample` to `.env` and update the ENV values
   - Download your `credentials.json` file from the Google Developers Console:

     - Go to https://console.developers.google.com
     - Go to **Credentials**
     - Download the application (OAuth client) JSON file
     - Save it as `credentials.json` _(or update `CREDENTIALS_FILE_NAME` ENV to_
       _match)_ to the project root

       - The following information will be pulled from `credentials.json`:

         - `client_id`
         - `client_secret`
         - `redirect_uris`

           _(if you have more than one, the project defaults to the first `uri`.)_

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

   - ```sh
     go run main.go
     ```

   - Go to `http://localhost:8000` in your browser

   - On your first run, you'll be prompted to authorize access:

     - You'll be prompted to sign in or select the account to use for
       authorization
     - Click `Accept`
     - Copy the code from the browser,\
       Paste it into the command-line prompt,\
       and press `Enter`.
     - Authorization info is stored in the file system,\
       the won't be prompted for authorization on the next run.
