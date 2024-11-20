Event Application Status Checker
================================

This Go script reads a CSV file containing event names and their corresponding URLs, fetches the content of each URL, and uses the OpenAI API to determine whether the event is currently accepting applications. It outputs the status for each event in the console.

Table of Contents
-----------------

*   [Prerequisites](#prerequisites)
*   [Installation](#installation)
*   [Setup](#setup)
*   [Usage](#usage)
*   [CSV File Format](#csv-file-format)
*   [Environment Variables](#environment-variables)
*   [License](#license)

Prerequisites
-------------

*   **Go 1.16 or later**: Ensure that Go is installed on your system. If not, download it from the [official website](https://golang.org/dl/).
*   **OpenAI API Key**: You need a valid OpenAI API key to use the OpenAI Chat Completion API.
*   **Internet Connection**: The script fetches web pages and communicates with the OpenAI API over the internet.

Installation
------------

1.  **Clone the Repository**
    
    ```bash
    git clone https://github.com/yourusername/event-status-checker.git
    cd event-status-checker
    ```
    
2.  **Install Dependencies**
    
    Use `go get` to install the required packages:
    
    ```bash
    go get github.com/joho/godotenv
    ```
    

Setup
-----

1.  **Create a `.env` File**
    
    In the project root directory, create a file named `.env` and add your OpenAI API key:
    
    ```dotenv
    OPENAI_API_KEY=your_openai_api_key_here
    ```
    
2.  **Prepare the CSV File**
    
    Ensure you have a `urls.csv` file in the project root directory. Refer to the [CSV File Format](#csv-file-format) section for details.
    

Usage
-----

Run the script using the following command:

```bash
go run main.go
```

The script will process each event in the `urls.csv` file and output whether the event is currently accepting applications.

CSV File Format
---------------

The `urls.csv` file should be a CSV file with the following headers:

*   **name**: The name of the event.
*   **url**: The URL of the event's webpage.

### Example

```csv
name,url
Event One,https://www.eventone.com
Event Two,https://www.eventtwo.com
```

Environment Variables
---------------------

The script uses the `OPENAI_API_KEY` environment variable for authentication with the OpenAI API. This is loaded from the `.env` file using the `godotenv` package.

### Example `.env` File

```dotenv
OPENAI_API_KEY=your_openai_api_key_here
```

How It Works
------------

1.  **Load Environment Variables**
    
    The script starts by loading the `.env` file to retrieve the OpenAI API key.
    
2.  **Read the CSV File**
    
    It opens `urls.csv` and reads each record, expecting `name` and `url` columns.
    
3.  **Fetch Web Page Content**
    
    For each URL, the script fetches the page content.
    
4.  **Query OpenAI API**
    
    It sends the page content to the OpenAI Chat Completion API with the prompt:
    
    > Based on the following webpage content, determine if the event is currently accepting applications. Answer only with 'yes' or 'no'.
    
5.  **Display Results**
    
    The script outputs whether each event is currently accepting applications based on the API's response.
    

Error Handling
--------------

*   **Missing Environment Variables**: The script will terminate if the `OPENAI_API_KEY` is not set.
*   **CSV File Issues**: It checks for the presence of the required columns and handles empty files.
*   **HTTP Errors**: Handles failures in fetching URLs and API requests gracefully.
*   **API Response Errors**: Validates responses from the OpenAI API and handles unexpected results.

License
-------

This project is licensed under the MIT License.