package main

import (
	"encoding/json"
	"fmt"
	"github.com/wenlng/go-captcha/captcha"
	"go-captcha-example/tools"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Example: Get captcha data
	http.HandleFunc("/go_captcha_data", getCaptchaData)
	// Example: Post check data
	http.HandleFunc("/go_captcha_check_data", checkCaptcha)
	// Example: demo
	http.HandleFunc("/go_captcha_demo", demo)

	// 临时定时清空缓存，由于是demo即在程序内部实现
	runTimedTask()

	log.Println("ListenAndServe 0.0.0.0:8001")
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Fatal("ListenAndServe err: ", err)
	}
}

// =========================================================

/**
 * @Description: demo
 * @param w
 * @param r
 */
func demo(w http.ResponseWriter, r *http.Request) {
	sessid := time.Now().UnixNano() / 1e6
	t, _ := template.ParseFiles(getPWD() + "/view/demo.html")
	_ = t.Execute(w, map[string]interface{}{"sessid": sessid})
}

/**
 * @Description: Example
 * @param w
 * @param r
 */
func getCaptchaData(w http.ResponseWriter, r *http.Request) {
	capt := captcha.GetCaptcha()

	//chars := "abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	//_ = capt.SetRangChars(strings.Split(chars, ""))

	//chars := []string{"HE","CA","WO","NE","HT","IE","PG","GI","CH","CO","DA"}
	//_ = capt.SetRangChars(chars)

	//chars := []string{"你","好","呀","这","是","点","击","验","证","码","哟"}
	//_ = capt.SetRangChars(chars)

	capt.SetTextRangFontColors([]string{
		"#006600",
		"#005db9",
		"#aa002a",
		"#875400",
		"#6e3700",
		"#333333",
		"#660033",
	})

	// capt.SetFont([]string{
	// 	getPWD() + "/resources/fonts/fzshengsksjw_cu.ttf",
	// 	getPWD() + "/resources/fonts/fzssksxl.ttf",
	// 	getPWD() + "/resources/fonts/hyrunyuan.ttf",
	// })

	// capt.SetBackground([]string{
	// 	getPWD() + "/resources/images/1.jpg",
	// 	getPWD() + "/resources/images/2.jpg",
	// 	getPWD() + "/resources/images/3.jpg",
	// 	getPWD() + "/resources/images/4.jpg",
	// 	getPWD() + "/resources/images/5.jpg",
	// })

	//capt.SetThumbBackground([]string{
	//	getPWD() + "/resources/images/thumb/r1.jpg",
	//	getPWD() + "/resources/images/thumb/r2.jpg",
	//	getPWD() + "/resources/images/thumb/r3.jpg",
	//	getPWD() + "/resources/images/thumb/r4.jpg",
	//	getPWD() + "/resources/images/thumb/r5.jpg",
	//})

	//capt.SetThumbBgCirclesNum(200)
	//capt.SetImageFontAlpha(0.5)

	dots, b64, tb64, key, err := capt.Generate()
	if err != nil {
		bt, _ := json.Marshal(map[string]interface{}{
			"code": 1,
			"message": "GenCaptcha err",
		})
		_, _ = fmt.Fprintf(w, string(bt))
		return
	}
	writeCache(dots, key)
	bt, _ := json.Marshal(map[string]interface{}{
		"code": 0,
		"image_base64": b64,
		"thumb_base64": tb64,
		"captcha_key": key,
	})
	_, _ = fmt.Fprintf(w, string(bt))
}

/**
 * @Description: Verify where the user clicks on the image
 * @param w
 * @param r
 */
func checkCaptcha(w http.ResponseWriter, r *http.Request) {
	code := 1
	_ = r.ParseForm()
	dots := r.Form.Get("dots")
	key := r.Form.Get("key")
	if dots == "" || key == "" {
		bt, _ := json.Marshal(map[string]interface{}{
			"code": code,
			"message": "dots or key param is empty",
		})
		_, _ = fmt.Fprintf(w, string(bt))
		return
	}

	cacheData := readCache(key)
	if cacheData == "" {
		bt, _ := json.Marshal(map[string]interface{}{
			"code": code,
			"message": "Invalid or expired key",
		})
		_, _ = fmt.Fprintf(w, string(bt))
		return
	}
	src := strings.Split(dots, ",")

	var dct map[int]captcha.CharDot
	if err := json.Unmarshal([]byte(cacheData), &dct); err != nil {
		bt, _ := json.Marshal(map[string]interface{}{
			"code": code,
			"message": "Invalid or expired key",
		})
		_, _ = fmt.Fprintf(w, string(bt))
		return
	}

	chkRet := false
	if len(src) >= len(dct) * 2 {
		chkRet = true
		for _, dot := range dct {
			i := dot.Index
			j := i * 2
			k := i * 2 + 1
			a, _ := strconv.Atoi(src[j])
			b, _ := strconv.Atoi(src[k])
			chkRet = checkDist(a, b, dot.Dx, dot.Dy, dot.Width, dot.Height)
			if !chkRet {
				break
			}
		}
	}

	if chkRet && (len(dct) * 2) == len(src) {
		code = 0
	}

	bt, _ := json.Marshal(map[string]interface{}{
		"code": code,
	})
	_, _ = fmt.Fprintf(w, string(bt))
	return
}

/**
 * @Description: Write Cache，“Redis” is recommended
 * @param v
 * @param file
 */
func writeCache(v interface{}, file string) {
	bt, _ := json.Marshal(v)
	month := time.Now().Month().String()
	cacheDir := getCacheDir() + month + "/"
	_ = os.MkdirAll(cacheDir, 0660)
	file = cacheDir + file + ".json"
	logFile, _ := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	defer logFile.Close()
	// 检查过期文件
	//checkCacheOvertimeFile()
	_, _ = io.WriteString(logFile, string(bt))
}

/**
 * @Description: Read Cache，“Redis” is recommended
 * @param file
 * @return string
 */
func readCache(file string) string {
	month := time.Now().Month().String()
	cacheDir := getCacheDir() + month + "/"
	file = cacheDir + file + ".json"

	if !checkFileIsExist(file) {
		return ""
	}

	bt, err := ioutil.ReadFile(file)
	err = os.Remove(file)
	if err == nil {
		return string(bt)
	}
	return ""
}

/**
 * @Description: Calculate the distance between two points
 * @param sx
 * @param sy
 * @param dx
 * @param dy
 * @param width
 * @param height
 * @return bool
 */
func checkDist(sx, sy, dx, dy, width int, height int) bool {
	return sx >= dx &&
		sx <= dx + width &&
		sy <= dy &&
		sy >= dy - height
}

/**
 * @Description: Get cache dir path
 * @return string
 */
func getCacheDir() string  {
	return getPWD() + "/.cache/"
}


/**
 * @Description: Get pwd dir path
 * @return string
 */
func getPWD() string {
	path, err := os.Getwd()
	if err != nil {
		return ""
	}
	return path
}

/**
 * @Description: Check file exist
 * @param filename
 * @return bool
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

/**
 * @Description: 启动定时任务, 5分钟执行一次
 */
func runTimedTask()  {
	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for range ticker.C {
			checkCacheOvertimeFile()
		}
	}()
}

/**
 * @Description: 检查缓存超时文件， 30分钟
 */
func checkCacheOvertimeFile()  {
	files, files1, _ := listDir(getCacheDir())
	for _, table := range files1 {
		temp,_,_ := listDir(table)
		for _,temp1 := range temp{
			files = append(files, temp1)
		}
	}

	for _, file := range files {
		t := tools.GetFileCreateTime(file)
		ex := time.Now().Unix() - t
		if ex > (60 * 30) {
			_ = os.Remove(file)
		}
	}
}

/**
 * @Description: 获取目录文件列表
 * @param dirPth
 * @return files
 * @return files1
 * @return err
 */
func listDir(dirPth string) (files []string,files1 []string, err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil,nil, err
	}

	PthSep := string(os.PathSeparator)
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			files1 = append(files1, dirPth+PthSep+fi.Name())
			_, _, _ = listDir(dirPth + PthSep + fi.Name())
		} else {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, files1, nil
}

