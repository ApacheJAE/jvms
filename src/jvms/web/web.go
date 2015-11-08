package web

import(
  "fmt"
  "net/http"
  "net/url"
  "os"
  "io"
  "io/ioutil"
  "strings"
  "strconv"
  "../arch"
  "../file"
)

var client = &http.Client{}

func SetProxy(p string){
  if p != "" && p != "none" {
    proxyUrl, _ := url.Parse(p)
    client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}
  } else {
    client = &http.Client{}
  }
}

func Download(url string, target string) bool {

  output, err := os.Create(target)
  if err != nil {
    fmt.Println("Error while creating", target, "-", err)
  }
  defer output.Close()

  response, err := client.Get(url)
  if err != nil {
    fmt.Println("Error while downloading", url, "-", err)
  }
  defer response.Body.Close()

  _, err = io.Copy(output, response.Body)
  if err != nil {
    fmt.Println("Error while downloading", url, "-", err)
  }

  if response.Status[0:3] != "200" {
    fmt.Println("Download failed. Rolling Back.")
    err := os.Remove(target)
    if err != nil {
      fmt.Println("Rollback failed.",err)
    }
    return false
  }

  return true
}

func GetJDK(root string, v string, url string, a string) bool {

  a = arch.Validate(a)

  if url == "" {
    //No url should mean this version/arch isn't available
    fmt.Println("JDK v"+v+" " + a + "bit isn't available right now.")
  } else {
    //jdk1.7.0_67.zip
    fileName := root + "\\download\\" + file.GenJDKZipFileName(v,a)

    fmt.Printf("Downloading jdk version "+v+" ("+a+"-bit)... ")

    if Download(url,fileName) {
      fmt.Printf("Complete\n")
      return true
    } else {
      return false
    }
  }
  return false

}

func GetNpm(root string, v string) bool {
  url := "https://github.com/npm/npm/archive/v"+v+".zip"
  // temp directory to download the .zip file
  tempDir := root+"\\temp"

  // if the temp directory doesn't exist, create it
  if (!file.Exists(tempDir)) {
    fmt.Println("Creating "+tempDir+"\n")
    err := os.Mkdir(tempDir, os.ModePerm)
    if err != nil {
      fmt.Println(err)
      os.Exit(1)
    }
  }
  fileName := tempDir+"\\"+"npm-v"+v+".zip"

  fmt.Printf("Downloading npm version "+v+"... ")
  if Download(url,fileName) {
    fmt.Printf("Complete\n")
    return true
  } else {
    return false
  }
}

func GetRemoteTextFile(url string) string {
  response, httperr := client.Get(url)
  if httperr != nil {
    fmt.Println("\nCould not retrieve "+url+".\n\n")
    fmt.Printf("%s", httperr)
    os.Exit(1)
  } else {
    defer response.Body.Close()
    contents, readerr := ioutil.ReadAll(response.Body)
    if readerr != nil {
      fmt.Printf("%s", readerr)
      os.Exit(1)
    }
    return string(contents)
  }
  os.Exit(1)
  return ""
}

func IsNode64bitAvailable(v string) bool {
  if v == "latest" {
    return true
  }

  // Anything below version 8 doesn't have a 64 bit version
  vers := strings.Fields(strings.Replace(v,"."," ",-1))
  main, _ := strconv.ParseInt(vers[0],0,0)
  minor, _ := strconv.ParseInt(vers[1],0,0)
  if main == 0 && minor < 8 {
    return false
  }

  // TODO: fixme. Assume a 64 bit version exists
  return true
}

func getNodeUrl (v string,  vpre string) string {
  url := "http://nodejs.org/dist/v"+v+"/" + vpre + "/node.exe"
  // Check online to see if a 64 bit version exists
  res, err := client.Head( url )
  if err != nil {
    return ""
  }
  if res.StatusCode == 200 {
    return url
  } else {
    return ""
  }
}
