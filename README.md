# webcrawler

A simple web crawler that deep parses the input url and all the nested urls on each page to extract linked URLs.

It supports a stopping mechanism to crawling, which is based on the max number of urls parsed. It is optional and when not provided, the crawler runs recursively as long as it doesn't encounter any http errors, dead ends or memory limit for this application execution.

## Input format:
```
$ ./go run main.go <initial/starting point url> <stopping limit>
```

## Pre-requisites to run this program locally
Need to have latest version of go lang installed
Active internet connection for crawling the URL

## Testing
Run the following command to test if this program is working as expected:
```
$ git clone git@github.com:ShreyasChikkabilathiHarisha/webcrawler.git
$ cd <path to the cloned repo>
$ go run main.go http://www.rescale.com <integer value for stopping point based on max number of urls crawled>
```

## Validation
The above command should print each URL and all its sub URLs, followed by the next URL and all of its sub URLs and so on. The parent URL should be printed first and all its sub urls should be printed with a tab space before on each line.

eg:
```
$ go run main.go http://www.rescale.com 3
Starting crawling from the initial URL:  http://www.rescale.com

http://www.rescale.com
	 https://info.rescale.com/case-studies/boom-supersonic
	 https://resources.rescale.com/news/
	 https://twitter.com/rescaleinc
	 https://www.facebook.com/rescaleinc/
         .
         .
https://info.rescale.com/case-studies/boom-supersonic
         https://info.rescale.com/case-studies/boom-supersonic/something
         .
         .
https://resources.rescale.com/news/
         https://resources.rescale.com/news/something
         .
         .
```

## Potential improvements that can immediately improve this current implementation
I have added comments in the code on few things that can be improved for quick performance benifits.

Few other improvements are with error handling/logging, filter on urls and waits/retries of http get call. There are times when the crawling stops due to few of the above mentioned issues, which can be significantly improved with follow-up patches.

## Validation
There is an inbuilt validation logic to make sure if the webcrawler is working as expected. The command to validate webcrawler is as follows:
```
go run main.go validate
```

The validation logic is based on the comparison of webcrawler result of 5000 urls against a snapshot of the same stored on the repo, which has 5000 urls for a specific initital url. 

The logic uses a comparison threshold of 0.2 as the number of similar urls will vary vastly based on number of factors: 
   1. Response time of the websites being crawled
   2. Physical place that the validation is run
   3. local machine threads writing order to the crawler results
   4. Websites being unresponsive
   5. Rate limiting on websites

Due to above listed reasons, a lower comparison threshold has been chosen to validate the results of webcrawler.

## Note
There are few options added as comments which indicate alternate ways of doing things with explanation on the same
