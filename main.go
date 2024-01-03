package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
)

type PageRequest struct {
	Url  string `json:"url"`
}

type PageResponse struct {
	ImageUrl     []string     `json:"image_url"`
	ChapterPager []PageNumber `json:"chapterPager"`
	PagePager    []PageNumber `json:"pagePager"`
	LongPage     bool         `json:"longPage"`
	CurrentPage  string       `json:"currentPage"`
}

type PageNumber struct {
	Number string
	Url    string
}

func main() {
	router := gin.Default()
	router.Use(CORSMiddleware())
	router.GET("/home", home)
	router.POST("/initial", initChapter)
	router.POST("/scrape", scrape)

	router.Run("localhost:8080")
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func home(c *gin.Context) {

}

func initChapter(c *gin.Context) {
	pr := PageRequest{Url: ""}

	if err := c.ShouldBindJSON(&pr); err != nil {
		// Handle error (e.g., invalid JSON format)
		c.JSON(http.StatusBadRequest, gin.H{"400 error": err.Error()})
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new browser session
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	response := pr.scrapeLongPage(ctx)
	fmt.Println("length ", len(response.ImageUrl))

	if len(response.ImageUrl) < 1 {
		response = pr.scrapePage(ctx)
		response.getPagePager(ctx)
	} else {
		response.getChapterPager(ctx)
		response.LongPage = true
	}
	c.IndentedJSON(http.StatusOK, response)
}

func scrape(c *gin.Context) {
	pr := PageRequest{Url: ""}

	if err := c.ShouldBindJSON(&pr); err != nil {
		// Handle error (e.g., invalid JSON format)
		c.JSON(http.StatusBadRequest, gin.H{"400 error": err.Error()})
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a new browser session
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	response := pr.scrapePage(ctx)
	response.getChapterPager(ctx)
	response.getPagePager(ctx)
	c.IndentedJSON(http.StatusOK, response)
}

func (pr *PageRequest) scrapeLongPage(ctx context.Context) (response PageResponse) {
	var nodes []*cdp.Node

	err := chromedp.Run(ctx, chromedp.Navigate(pr.Url))
	if err != nil {
		log.Fatal(err)
	}

	longpageCtx, loginCancel := context.WithTimeout(ctx, 1*time.Second)
	defer loginCancel()

	err2 := chromedp.Run(longpageCtx,
		chromedp.WaitVisible(`.load-src`, chromedp.ByQuery),
		// chromedp.Query(`.load-src`, chromedp.ByQuery),
		chromedp.Nodes(`.load-src`, &nodes, chromedp.ByQueryAll),
	)
	if err2 != nil {
		fmt.Println("Not long image")
	}

	// print titles
	for _, node := range nodes {
		response.ImageUrl = append(response.ImageUrl, node.AttributeValue("src"))
	}

	return response
}

func (pr *PageRequest) scrapePage(ctx context.Context) (response PageResponse) {
	var nodeValue string
	var ok bool

	err3 := chromedp.Run(ctx,
		chromedp.Navigate(pr.Url),
		chromedp.WaitVisible(`#cp_image`),
		chromedp.AttributeValue(`#cp_image`, "src", &nodeValue, &ok),
	)
	if err3 != nil {
		fmt.Println("Images not found")
	}

	if ok {
		// fmt.Println("Image found: ", nodeValue)
		response.ImageUrl = append(response.ImageUrl, nodeValue)
	}
	fmt.Println("Image Url: ", response.ImageUrl)
	return response
}

func (response *PageResponse) getChapterPager(ctx context.Context) {
	var chapterArray []PageNumber

	var nodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Nodes(`.view-paging>div>a`, &nodes, chromedp.ByQueryAll),
	)
	if err != nil {
		log.Fatal(err)
	}

	/* Get prev chapter and  */
	// divide by 2 first
	nodes = nodes[0 : len(nodes)/2]
	for _, node := range nodes {
		// if href starts with /, then it's a chapter
		fmt.Println(`node =`, node.AttributeValue("href"))
		if node.AttributeValue("href")[0] == '/' {
			chapterArray = append(chapterArray, PageNumber{Number: "", Url: node.AttributeValue("href")})
		}
	}
	response.ChapterPager = chapterArray
}

func (response *PageResponse) getPagePager(ctx context.Context) {
	var pagesArray []PageNumber
	// var pagesArray []string
	var nodes []*cdp.Node
	// var pageNumber []string
	selector := "#chapterpager>a"
	var pageNumbers []string
	var currentPage string

	/* Get pages */
	err2 := chromedp.Run(ctx,
		chromedp.Nodes(`#chapterpager>a`, &nodes, chromedp.ByQueryAll),
		// get all page number
		chromedp.Evaluate(fmt.Sprintf(`
		var elements = document.querySelectorAll('%s');
		var htmlArray = [];
		for (var i = 0; i < elements.length; i++) {
			htmlArray.push(elements[i].innerHTML);
		}
		htmlArray;
		`, selector), &pageNumbers),
		chromedp.Evaluate(`document.querySelector('.current').innerHTML;`, &currentPage),
	)
	if err2 != nil {
		log.Fatal(err2)
	}
	pageNumbers = pageNumbers[0 : len(pageNumbers)/2]

	// divide pages by 2 first
	nodes = nodes[0 : len(nodes)/2]
	for index, node := range nodes {
		fmt.Println(`node =`, node.AttributeValue("href"))
		pagesArray = append(pagesArray, PageNumber{Number: pageNumbers[index], Url: node.AttributeValue("href")})
	}
	response.PagePager = pagesArray
	response.CurrentPage = currentPage
}
