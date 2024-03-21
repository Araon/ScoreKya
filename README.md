
# ScoreKya?
<p align="center">
   <img src="docs/logo.png" alt="Cricket Logo" width="100" height="100">
</p>

This Go application fetches live cricket scores from Cricbuzz and generates AI-powered meta commentary using OpenAI's GPT-3.5 model.

## Features

- Retrieves live cricket scores and player statistics from Cricbuzz.
- Enables optional AI-generated meta commentary for matches.
- Uses concurrent scraping and API requests for efficiency.
- Supports dynamic selection of matches via CLI menu.
- Integrates OpenAI's GPT-3.5 for generating cricket commentary.

## Usage

1. Clone the repository and navigate to the project directory.
2. Set up your environment variables by creating a `.env` file with the following contents:
```
OPENAI_API_KEY=your_openai_api_key_here
```

3. Build and run the application using the following command:
```bash
go run main.go

Dependencies

    Colly: For web scraping Cricbuzz.
    Godotenv: For loading environment variables from a .env file.
    gocliselect: For building CLI menus.
    go-openai: For interacting with OpenAI's GPT-3.5 API.
```
## Example

Here's an example of how to use the application:

# screenshot here

## License

This project is licensed under the MIT License. See the LICENSE file for details.

