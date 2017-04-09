package main

import (
	"fmt"
	"log"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
	"regexp"
	"os"
	"net/http"
)

var path = "info.txt"

func handler(w http.ResponseWriter, req *http.Request) {
	sess := session.Must(session.NewSession())

	nameFilter := "best"
	awsRegion := "eu-central-1"
	svc := ec2.New(sess, &aws.Config{Region: aws.String(awsRegion)})
	fmt.Printf("listing instances with tag %v in: %v\n", nameFilter, awsRegion)
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:test"),
				Values: []*string{
					aws.String(strings.Join([]string{"*", nameFilter, "*"}, "")),
				},
			},
		},
	}
	resp, err := svc.DescribeInstances(params)
	re := regexp.MustCompile(`PublicIpAddress:(.*)`)
	text := fmt.Sprintf("%v\n", *resp)
	matches := re.FindAllString(text, -1)
	r := strings.NewReplacer("PublicIpAddress: \"", "", "\",", "")

	if err != nil {
		fmt.Println("there was an error listing instances in", awsRegion, err.Error())
		log.Fatal(err.Error())
	}
	deleteFile()
	createFile()
	var file, err_open = os.OpenFile(path, os.O_WRONLY | os.O_CREATE | os.O_APPEND, 0777)
	checkError(err_open)
	defer file.Close()

	for _,element := range matches {
		ip := r.Replace(element)
		str := fmt.Sprintf(
			`$fw tcp 8091:8094 %[1]v
$fw tcp 11207:11211 %[1]v
$fw tcp 18091:18093 %[1]v
$dck tcp 6000 %[1]v`, ip)

		_, erroring := file.WriteString(str+"\n")
		checkError(erroring)
	}
	file.Sync()
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}
func createFile() {
	// detect if file exists
	var _, err = os.Stat(path)

	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		checkError(err)
		defer file.Close()
	}
}
func deleteFile() {
	// delete file
	var _, err = os.Stat(path)

	// create file if not exists
	if err == nil {
		var err = os.Remove(path)
		checkError(err)
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
