package main

import (
	"strings"
	"fmt"
	"os"
	"net/http"
	"io"
	"log"
	"archive/zip"
	"bufio"
	"strconv"
	"math"
	"bytes"
)

func download(url string) {
	tokens := strings.Split(url, "/");
	filename := tokens[len(tokens) - 1]
	output, error := os.Create(filename)
	if error != nil {

		return;
	}
	defer output.Close()

	response, err := http.Get(url)

	if err != nil {
		return;
	}

	defer response.Body.Close()

	bytesDownloaded , err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println(" Download error ")
	}

	fmt.Println(bytesDownloaded, " bytes Downloaded ")
	pwd, err := os.Getwd();
	if err != nil {
		return
	}

	fp := pwd + string(os.PathSeparator) + filename
	r, err := zip.OpenReader(fp)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
	/*	fmt.Println("Contents of %s:\n", f.Name)*/
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		extractedfile := "extracted.csv"
		etractedoutput, error := os.Create(extractedfile)
		defer etractedoutput.Close();
		if error != nil {
			log.Fatal(error)
		}

		_, err = io.Copy(etractedoutput, rc)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
		

		extractedReader , errror := os.Open(extractedfile)
		if errror != nil {
			log.Fatal(errror)

		}
		defer extractedReader.Close()

		wf,err  := os.Create("embargo.conf")
		defer wf.Close()
		
		if err != nil {
			log.Fatal(err)
		}

		n3, err := wf.WriteString("geo $http_x_forwarded_for $blacklist {\n")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("wrote %d bytes\n", n3)
		
		wf.WriteString("default 0;\r\n")
		

		scanner := bufio.NewScanner(extractedReader)
		for scanner.Scan() {
			str := scanner.Text()
			tokens := strings.Split(str, ",")
			country := tokens[4]

			// {"CU", "IR", "SY", "SD", "KP"}
			if  strings.Contains(country,"CU") || strings.Contains(country,"IR") || strings.Contains(country, "SY") ||
					strings.Contains(country, "SD") || strings.Contains(country, "KP") {
				fromip , toip := tokens[0], tokens[1]
				fromip = strings.Replace(fromip, "\"", "", -1)
				toip = strings.Replace(toip, "\"", "", -1)

				start := ipToLong(fromip)
				end := ipToLong(toip)
				res  :=  cidrCalculator(start, end)

				for i :=0 ;i<len(res) ;i++ {
					outpiut := fmt.Sprintln(strings.Trim(res[i]," "),"1; \r")
					wf.WriteString(outpiut)
				}
			}
		}
		wf.WriteString("}")
		wf.Sync()
		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}


		extractedReader.Close()
		errror = os.Remove(extractedfile)
		if errror !=nil {
			log.Fatal(errror)
			
		}
		
	}
}

func ipToLong(ip string) uint64 {
	var li[4] uint64
	spip := strings.Split(ip, ".")
	for k := 0 ; k < 4 ; k++ {
		i, err := strconv.ParseUint(spip[k], 0, 64)
		if err != nil {
			log.Fatal(err)
		}
		li[k] = i
	}
	return (li[0] << 24) + (li[1] << 16) + (li[2] << 8) + li[3]
}

func cidrCalculator(start uint64, end uint64) []string {
	var result []string;
	cidrmask := []int { 0x00000000, 0x80000000,
		0xC0000000, 0xE0000000, 0xF0000000, 0xF8000000, 0xFC000000,
		0xFE000000, 0xFF000000, 0xFF800000, 0xFFC00000, 0xFFE00000,
		0xFFF00000, 0xFFF80000, 0xFFFC0000, 0xFFFE0000, 0xFFFF0000,
		0xFFFF8000, 0xFFFFC000, 0xFFFFE000, 0xFFFFF000, 0xFFFFF800,
		0xFFFFFC00, 0xFFFFFE00, 0xFFFFFF00, 0xFFFFFF80, 0xFFFFFFC0,
		0xFFFFFFE0, 0xFFFFFFF0, 0xFFFFFFF8, 0xFFFFFFFC, 0xFFFFFFFE,
		0xFFFFFFFF }
	for ; end >= start; {
		var maxsize int = 32
		for ; maxsize > 0; {
			var mask uint64 = uint64(cidrmask[maxsize - 1])
			var maskedBase uint64 = start & mask;
			if (maskedBase != start) {
				break;
			}
			maxsize--;
		}

		lg := (end - start + 1)
		var logof2 float64 = math.Ln2
		var endstart  float64 = float64(lg)
		var b float64 = math.Log(endstart)
		x := b / logof2

		var maxdiff int = int(32 - math.Floor(x))

		if (maxsize < maxdiff) {
			maxsize = maxdiff;
		}
		formattedIp := longtoIp(start)
		formattedIpsplit := strings.Split(formattedIp, "")
		formattedIpsplit = append(formattedIpsplit, "/")
		formattedIpsplit = append(formattedIpsplit, strconv.Itoa(maxsize))

		start= start + uint64(math.Pow(2,float64(32-maxsize)))
		var output  = strings.Join(formattedIpsplit, "")
		result = append(result,output)
	
	}
	return result
}


func longtoIp(ip uint64) string {
	var buffer bytes.Buffer
	var x string = strconv.FormatUint(ip>>24, 10)
	buffer.WriteString(x)
	buffer.WriteString(".")
	x = strconv.FormatUint((ip&0x00FFFFFF)>>16, 10)
	buffer.WriteString(x)
	buffer.WriteString(".")
	x = strconv.FormatUint((ip&0x0000FFFF)>>8, 10)
	buffer.WriteString(x)
	buffer.WriteString(".")
	x = strconv.FormatUint(ip&0x000000FF, 10);
	buffer.WriteString(x)
	return buffer.String()
}

func main() {
	url := "http://geolite.maxmind.com/download/geoip/database/GeoIPCountryCSV.zip"
	download(url)
}
