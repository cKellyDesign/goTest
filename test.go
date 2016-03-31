package main

import (
  "fmt"
  "encoding/json"  
  "io/ioutil"
  "strings"
  "net/http"
  "reflect"
  "github.com/plimble/ace" // Router Library
  "gopkg.in/olivere/elastic.v3" // ElasticSearch Client Library
)

type hitStruct struct {
  _id string
  _type string
  found bool
  _source *json.RawMessage
}



func main() {
  // Create New Router
  a := ace.New()

  // Create new ElasticSearch Client
  client, err := elastic.NewClient()
  if err != nil {
    panic(err)
  }

  // Test new ElasticSearch Client at default "http://127.0.0.1:9200"
  info, code, err := client.Ping("http://127.0.0.1:9200").Do()
  if err != nil {
    panic(err)
  }
  fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

  // Check the ElasticSearch Client for the "reports" root level index
  exists, err := client.IndexExists("reports").Do()
  if err != nil {
    panic(err)
  }

  // If ElasticSearch Client does not have root level index, create it
  if !exists {
    createIndex, err := client.CreateIndex("reports").Do()
    if err != nil {
      panic(err)
    }
    if !createIndex.Acknowledged {
      panic("CREATE INDEX NOT Acknowledged")
    }
  }

  // Root Route Handler
  a.GET("/", func (c *ace.C){
    c.String(200, "Hello World")
  })

  a.GET("/perfReport/:type", func(c *ace.C){
    reportType := c.Param("type")

    // fmt.Printf("\nHitting /perfReport/%s\n",reportType)

    // termQuery := elastic.NewTermQuery("type", reportType)
    termQuery := elastic.NewMatchAllQuery()
    searchResult, err := client.Search().
      Index("reports").
      Query(termQuery).
      Pretty(true).
      Do()
    if err != nil {
      panic(err)
    }

    fmt.Printf("\nHitting /perfReport/%s resulting in %d hits\n", reportType, searchResult.TotalHits())

    // fmt.Printf("\n%+v\n", searchResult.Hits.Hits)

    // if searchResult.Hits != nil {
    //   comboString := ""

      for _, hit := range searchResult.Hits.Hits {
        // fmt.Printf("\n%+v\n", hit.Source)
        fmt.Println(reflect.TypeOf(hit.Source))
      }


    //     var hitSourceString string
    //     data := (*json.RawMessage)(hit.Source)
    //     err := json.Unmarshal(*data, &hitSourceString)
    //     if err != nil {
    //       panic(err)
    //     }
    //     comboString := fmt.Sprintf("%s, %s", comboString, hitSourceString)
    //     fmt.Printf("\n%s\n", comboString)
    //   }

    //   toReturn := fmt.Sprintf("[%s]", comboString)


    //   c.String(200, toReturn)
    // } else {
    //   c.String(200, "No Results")
    // }

  })

  // Reposrt Posting Handler
  a.POST("/perfReport/:type/:sessionID", func(c *ace.C){

    reportType, sessionID := c.Param("type"), c.Param("sessionID")
    requestString := fmt.Sprintf("http://localhost:9200/reports/%s/%s", reportType, sessionID)
    fmt.Printf("\n\n%s\n\n", requestString)
    // Grab report data from POST body
    data, err := ioutil.ReadAll(c.Request.Body)
    if err != nil {
      panic(err)
    }

    check1, err := http.Get(requestString)
    if err != nil {
      panic(err)
    } else {
      defer check1.Body.Close()
      checkData, err2 := ioutil.ReadAll(check1.Body)
      if err2 != nil {
        panic(err)
      }

      ind := strings.Index(string(checkData), "\"found\":true")
      fmt.Printf("\n\nSubstring FOUND has index of %s", int(ind))
      fmt.Printf("\n\nData : %s", string(checkData))

      // b := []byte(checkData)
      // var repLog reportLog
      // err3 := json.Unmarshal(b, &repLog)
      // if err3 != nil {
      //   panic(err3)
      // }
      // rp := &repLog
      // fmt.Printf("\n\n%s report for%s  %s\n\n", rp._type, rp._id, rp.found)
    }

    // ElasticSearch Client GET Api
    // get1, err := client.Get().
    //   Index("reports").
    //   Type(reportType).
    //   Id(sessionID).
    //   Do()
    // if err != nil {
    //   panic(err)
    // }

    // if get1.Found {
    //   fmt.Printf("SESSION ID for %s already exists", reportType)
    // }

    // if stored != true; err != nil {
    //   panic("did not store value")
    // }

    // Add Report to ElasticSearch Client Index "reports"
    put1, err := client.Index().
      Index("reports").
      Type(reportType).
      Id(sessionID).
      BodyString(string(data)).
      Do()
    if err != nil {
      panic(err)
    }


    fmt.Printf("\n%s", put1)
    fmt.Printf("\n%s", requestString)
    c.String(200, requestString)
  })

  // Get Report data from ElasticSearch Client in form of JSON
  a.GET("/perfReport/:type/:sessionID", func(c *ace.C){
    reportType, sessionID := c.Param("type"), c.Param("sessionID")

    // ElasticSearch Client GET Api
    get1, err := client.Get().
      Index("reports").
      Type(reportType).
      Id(sessionID).
      Do()
    if err != nil {
      panic(err)
    }

    data := convertDataToString(get1.Source)

    fmt.Printf("\nData : %s", data)
    c.String(200, data)
  })

  // Run Router on Port 8080
  a.Run(":8080")
}

// Convert Raw JSON response to returnable string
func convertDataToString (src *json.RawMessage) string {
  data := (*json.RawMessage)(src)
  var toReturn string
  err := json.Unmarshal(*data, &toReturn)
  if err != nil {
    panic(err)
  }
  return toReturn
}






// func storeData (reportType, sessionID, data, cli) (bool, error) {

//   // check for existing sessionID first
//   // exists, err := getData()
//   // if exists, update
//   // updated, err := updateData()

//   stored, err := cli.Index().
//     Index("reports").
//     Type(reportType).
//     Id(sessionID).
//     BodyString(data).
//     Do()

//   if err != nil {
//     return false, err
//   }

//   return true, nil
// }

// func getData (reportType, sessionID, cli) (string, error) {
//   gotten, err := cli.Get().
//     Index("reports").
//     Type(reportType).
//     Id(sessionID).
//     Do()
//   if err != nil {
//     return nil, err
//   }  
//   return gotten, nil
// }

// func updateData (reportType, sessionID, cli) (bool, error) {

//   return true, nil
// }

