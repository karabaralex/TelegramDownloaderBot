package operations

import (
    "fmt"
	"time"
)

type OperationResult struct {
	Id int
	Text string
}

func Download(channel chan OperationResult, id int, url string, destination string) {
	start := time.Now()
	fmt.Println("Download...")
	error:=DownloadFile(destination,url)
	elapsed := time.Now().Sub(start)
	if error!=nil {
		fmt.Println("Error downloading ", error)
		channel<-OperationResult{Id:id,Text:"error"}
		return
	}
	fmt.Println("Download took ", elapsed)
	channel<-OperationResult{Id:id,Text:"scheduled"}
}
