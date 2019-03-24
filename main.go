package main
import(
	"os"
	"fmt"
	"time"
	"strings"
	"os/exec"
	"strconv"
	"net/http"
	"io/ioutil"
	"database/sql"
	"encoding/base64"
	_ "github.com/mattn/go-sqlite3"
)
var PortCount = 9000
func main(){
	fsh := http.FileServer(http.Dir("WWW/assets"))
    http.Handle("/assets/", http.StripPrefix("/assets/", fsh))
	http.HandleFunc("/Login",Login)
	http.HandleFunc("/New",New)
	http.HandleFunc("/Renew",Renew)
	http.HandleFunc("/Work",Work)
	http.HandleFunc("/Init",Init)
	http.HandleFunc("/",Index)
	fmt.Println("Instant VSCode Start Success...Listening Port 88...")
	if err := http.ListenAndServe(":88", nil); err != nil {
		fmt.Println("ListenAndServe:", err)
	}
}
func Index(w http.ResponseWriter, r *http.Request) {
	f, err := os.OpenFile("WWW/index.html", os.O_RDONLY,0600)
	defer f.Close()
	if err !=nil {
		w.Write([]byte(err.Error()))
		return
	} else {
		HTMLByte,err:=ioutil.ReadAll(f)
		if err != nil{
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(HTMLByte)
	}
}
func Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	UserID := r.FormValue("ID")
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil{
		fmt.Println(err)
		w.Write([]byte("服务器开小差了...请等等再来..."))
		return
	}
	rows, err := db.Query("SELECT * FROM info WHERE id='"+ UserID +"'")
	if err != nil{
		fmt.Println(err)
		w.Write([]byte("服务器开小差了...请等等再来..."))
		return
	}
	for rows.Next() {
        var ID string
		var password string
		var dir string
		var bakdir string
		var createtime string
		var livetime string
        err = rows.Scan(&ID, &password, &dir, &bakdir, &createtime, &livetime)
        if err != nil{
			fmt.Println(err)
			w.Write([]byte("服务器开小差了...请等等再来..."))
			return
		}
		//w.Write([]byte(ID + " " +password + " "+dir+" "+bakdir+" "+createtime+" "+livetime))
		//nowTime := time.Now()
		token := base64.StdEncoding.EncodeToString([]byte(dir))
		http.Redirect(w,r,"/Work?id="+ID+"&token="+token,302)
		return 
	}
	http.Redirect(w,r,"/New?id="+UserID,302)
	
}
func New(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	UserID := r.FormValue("id")
	f, err := os.OpenFile("WWW/new.html", os.O_RDONLY,0600)
	defer f.Close()
	if err !=nil {
		w.Write([]byte(err.Error()))
		return
	} else {
		HTMLByte,err:=ioutil.ReadAll(f)
		if err != nil{
			w.Write([]byte(err.Error()))
			return
		}
		HTML := string(HTMLByte)
		if UserID == ""{
			HTML = strings.Replace(HTML,"{{$value}}","",-1)
		} else {
			HTML = strings.Replace(HTML,"{{$value}}",UserID,-1)
		}
		w.Write([]byte(HTML))
	}
}
func Renew(w http.ResponseWriter, r *http.Request) {
	f, err := os.OpenFile("WWW/renew.html", os.O_RDONLY,0600)
	defer f.Close()
	if err !=nil {
		w.Write([]byte(err.Error()))
		return
	} else {
		HTMLByte,err:=ioutil.ReadAll(f)
		if err != nil{
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(HTMLByte)
	}
}
func Work(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	UserID := r.FormValue("id")
	token, err := base64.StdEncoding.DecodeString(r.FormValue("token"))
	if err != nil {
		w.Write([]byte("请输入正确的参数"))
		return
	}
	f, err := os.OpenFile("WWW/work.html", os.O_RDONLY,0600)
	defer f.Close()
	if err !=nil {
		w.Write([]byte(err.Error()))
		return
	} else {
		HTMLByte,err:=ioutil.ReadAll(f)
		if err != nil{
			w.Write([]byte(err.Error()))
			return
		}
		HTML := string(HTMLByte)
		if UserID == ""{
			w.Write([]byte("请输入正确的参数"))
			return
		} else {
			if string(token) == ""{
				w.Write([]byte("请输入正确的参数"))
				return
			} else {
				HTML = strings.Replace(HTML,"{{$id}}",UserID,-1)
				HTML = strings.Replace(HTML,"{{$url}}",string(token),-1)
			}
		}
		w.Write([]byte(HTML))
	}
}
func Init(w http.ResponseWriter, r *http.Request) {
	f, err := os.OpenFile("WWW/renew.html", os.O_RDONLY,0600)
	defer f.Close()
	if err !=nil {
		w.Write([]byte(err.Error()))
		return
	} else {
		HTMLByte,err:=ioutil.ReadAll(f)
		if err != nil{
			w.Write([]byte(err.Error()))
			return
		}
		r.ParseForm()
		UserID := r.FormValue("id")
		Password := r.FormValue("password")
		/*LiveTime, err := strconv.ParseInt(r.FormValue("time"), 10, 64)
		if err != nil{
			w.Write([]byte("服务器开小差了...请等等再来..."))
		}*/
		LiveTime := 24
		MainDir := "/home/ubuntu/code-server/public"
		os.Mkdir(MainDir+UserID,os.ModePerm)
		go NewWork(WorkThread,time.Second*time.Duration(LiveTime)*3600,string(PortCount),MainDir+UserID,Password,UserID)
		db, err := sql.Open("sqlite3", "data.db")
		if err != nil{
			fmt.Println(err)
			w.Write([]byte("服务器开小差了...请等等再来..."))
			return
		}
		stmt, err := db.Prepare("INSERT INTO info values(?,?,?,?,?,?)")
		if err != nil{
			w.Write([]byte("服务器开小差了...请等等再来..."))
			return
		}
		res, err := stmt.Exec(UserID,Password,MainDir+UserID,"null",strconv.FormatInt(time.Now().Unix(),10),string(LiveTime))
		if err != nil{
			w.Write([]byte("服务器开小差了...请等等再来..."))
			return
		}
		id, err := res.LastInsertId()
		if err != nil{
			w.Write([]byte("服务器开小差了...请等等再来..."))
			return
		}
		fmt.Println(id)
		PortCount = PortCount + 1
		token := base64.StdEncoding.EncodeToString([]byte("https://123.207.241.119:"+string(PortCount)))
		HTML := string(HTMLByte)
		HTML = strings.Replace(HTML,"{{$url}}","https://123.207.241.119:"+string(PortCount)+"/Work?id="+UserID+"&token="+token,-1)
		w.Write([]byte(HTML))
		//http.Redirect(w,r,"/Work?id="+UserID+"&token="+token,302)

	}
}
func WorkThread(port string,dir string,passwd string){
	params := []string{"/home/ubuntu/code-server/code-server","-p="+port,"-d="+dir,"--password="+passwd}
	cmd := exec.Command("sudo", params...)
	fmt.Println(cmd.Args)
	cmd.Start()
	cmd.Wait()
}
func NewWork(workthread func(port string,dir string,passwd string),t time.Duration,port string,dir string,passwd string,UserID string){
	id:=make(chan int)
    ok:=make(chan int)
    go func(){
        id<-os.Getpid()
        workthread(port,dir,passwd)
        ok<-1
    }()
    proc:=os.Process{Pid:<-id}
    select{
        case <-time.After(t):
			proc.Kill()
			db,_ := sql.Open("sqlite3", "data.db")
			stmt,_ := db.Prepare("delete from info where id=?")
			stmt.Exec(UserID)
        case <-ok:
    }
}