package main

import (
  "fmt"
  "encoding/json"  
  "io/ioutil"
  // "strings"
  // "net/http"
  // "reflect"
  "github.com/plimble/ace" // Router Library
  "gopkg.in/olivere/elastic.v3" // ElasticSearch Client Library
)

type SearchHit struct {
  ID string
  ReportType string
  Found bool
  Source *json.RawMessage
}

type IndexQuery struct {
  EnvData []*json.RawMessage
  AdData []*json.RawMessage
  AssetData []*json.RawMessage
  EventLogs []*json.RawMessage
}

type ReportPost struct {
  ID          string
  ReportType  string
  Data        []string
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

    termQuery := elastic.NewMatchAllQuery()
    searchResults, err := client.Search().
      Index("reports").
      Query(termQuery).
      Pretty(true).
      Do()
    if err != nil {
      panic(err)
    }

    // thisIndex := &IndexQuery{}
    // for _, hit := range searchResults.Hits.Hits {

    //   switch hit.Type {
    //   case "assetData" :
    //     thisIndex.AssetData = append(thisIndex.AssetData, hit.Source)

    //   case "envData" :
    //     thisIndex.EnvData = append(thisIndex.EnvData, hit.Source)

    //   case "adData" :
    //     thisIndex.AdData = append(thisIndex.AdData, hit.Source)

    //   case "eventLogs" :
    //     thisIndex.EventLogs = append(thisIndex.EventLogs, hit.Source)
    //   }

    // }

    dataString := fmt.Sprintf("%s\n\nThere are <strong>%d</strong> entries in the Elasticsearch Database.", "<h1>Go Performance Layer</h1>", searchResults.TotalHits())
    c.String(200, dataString)
  })

  // Report Posting Handler
  a.POST("/perfReport/:type/:sessionID", func(c *ace.C){
    // CREATE UNIQUE 


    reportType, sessionID := c.Param("type"), c.Param("sessionID")
    requestString := fmt.Sprintf("http://localhost:8080/perfReport/%s/%s", reportType, sessionID)
    fmt.Printf("\n\n%s\n\n", requestString)

    // Grab report data from POST body
    data, err := ioutil.ReadAll(c.Request.Body)
    if err != nil {
      panic(err)
    }

    reportStruct := &ReportPost{ID: sessionID, ReportType: reportType}
    
    // Successfully check if the report type works to either create new document or update existing one
    exists, err := ExistsDocByID(reportStruct, client)
    if err != nil {
      panic(err)
    }
    fmt.Printf("\n%s exists = %d\n", sessionID, exists)

    if exists {
      // update document
      updatedDoc, err := UpdateDocByID(reportStruct, string(data), client)
      if err != nil {
        panic(err)
      }

      fmt.Printf("\n Updated Doc : %s", updatedDoc)
    } else {
      // create new document
      reportStruct.Data = append(reportStruct.Data, string(data))
      indexedDoc, err := IndexDocByID(reportStruct, client)
      if err != nil {
        panic(err)
      }

      fmt.Printf("\nindexedDoc: %s", indexedDoc)


      // oldRawData := (*json.RawMessage)(indexedDoc.Source)
      // var oldData string
      // mErr := json.Unmarshal(*oldRawData, &oldData)
      // if mErr != nil {
      //   panic(mErr)
      // }
      // fmt.Printf("*json.RawMessage to string!!! %s", oldData)


    }

    fmt.Printf("DATA : %s\n\n", string(data))
    
    c.String(200, string(data))
  })

  a.GET("/perfReport/:type/:sessionID", func(c *ace.C){
    reportStruct := &ReportPost{ID: c.Param("sessionID"), ReportType: c.Param("type")}

    exists, err := ExistsDocByID(reportStruct, client)
    if err != nil {
      panic(err)
    }
    if !exists {
      c.String(404, "This document does not exist.")
      return
    }

    doc, err := GetDocByID(reportStruct, client)
    if err != nil {
      panic(err)
    }
    
    c.JSON(200, doc.Source)
  })

  // Run Router on Port 8080
  a.Run(":8080")
}


func IndexDocByID (reportPost *ReportPost, client *elastic.Client) (doc *elastic.IndexResponse, err error) {
  doc, err = client.Index().
    Index("reports").
    Type(reportPost.ReportType).
    Id(reportPost.ID).
    BodyString(reportPost.Data[0]).
    Do()

  return
}



func UpdateDocByID (reportPost *ReportPost, data string, client *elastic.Client) (doc *elastic.IndexResponse, err error) {
  existingDoc, err := GetDocByID(reportPost, client)
  if err != nil {
    fmt.Printf("\n\n~~~ Update - doc doesn't exit\n\n")
    return
  }

  // convert *json.RawMessage to string
  oldRawData := (*json.RawMessage)(existingDoc.Source)
  var topData interface{}
  // var oldData string
  err = json.Unmarshal(*oldRawData, &topData)
  if err != nil {
    fmt.Printf("\n\n~~~ Update - couldn't Unmarshal\n\n")
    return
  }
  newestString, err := json.Marshal(topData)
  if err != nil {
    fmt.Printf("\n\n~~~ Update - newestString marshal failed")
    return
  }
  fmt.Printf("\n\n~~~~ Update - Unmarshalled into Interface{} - %v\n\n", topData)

  // secondString := json.Marshal(topData)

  // oldDataString := strings.TrimSuffix(oldData, "]}")
  // newDataString := strings.Replace(data, `{"data":[`, "", 1)

  // dataString := []string{oldDataString, newDataString}
  // bodyDataString := strings.Join(dataString, ", ")

  // fmt.Printf("\n\nAPPENDED STRING : %s\n\n", bodyDataString)

  // dataString := string(data)
  // dataString = strings.Replace(dataString, `{"data":[`, "", 1)
  // dataString = strings.TrimSuffix(dataString, "]}")

  // var partial interface{}


  doc, err = client.Index().
    Index("reports").
    Type(reportPost.ReportType).
    Id(reportPost.ID).
    BodyString(string(newestString)).
    Do()

  return
}

func GetDocByID (reportPost *ReportPost, client *elastic.Client) (doc *elastic.GetResult, err error) {
  doc, err = client.Get().
    Index("reports").
    Type(reportPost.ReportType).
    // Id(reportPost.ID).
    Do()

  return
}

func ExistsDocByID (reportPost *ReportPost, client *elastic.Client) (bool, error) {
  exists, err := client.Exists().Index("reports").Type(reportPost.ReportType).Id(reportPost.ID).Do()
  if !exists {
    return false, nil
  } 
  if err != nil {
    return false, err
  }
  return true, nil
}

func JoinDatas (oldData map[string]byte, newData map[string]byte) (combinedData []byte) {


  return
}

func convertDataToString (src *json.RawMessage) string {
  data := (*json.RawMessage)(src)
  var toReturn string
  err := json.Unmarshal(*data, &toReturn)
  if err != nil {
    panic(err)
  }
  return toReturn
}