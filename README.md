# httpecker
Takes a list of http or https urls and returns their status codes

## Installation
`go get github.com/Gaxpert/httpecker`

## Usage
Default usage checks for http and https status code.

It can take a list of urls from a file 

`httpecker -f url_list.txt`

or standard input

`cat url_list.txt | httpecker`

To check only http or https supply the flag. If any urls don't have the prefix, or they are not the right protocol (an http url used with https flag) they will be modified to fit the parameter

`cat url_list.txt | httpecker --http-only`

`cat url_list.txt | httpecker --https-only`

Default threads are 5, to change the value

`httpecker -f url_list.txt -t 50`

