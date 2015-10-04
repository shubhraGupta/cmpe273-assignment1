package main

import ("encoding/json"
	"errors"
	"fmt"
	"net/http"
	"io/ioutil"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strconv"
	"strings"
	"os"
)

var Id int

type CalcResp struct {
	TradeId int
	StocksInfo string
	Balance float64
}

var cr [2000]CalcResp

type PortfolioResp struct {
	StocksInfo string
	CurMarketVal float64
	Balance float64
}


type Input struct{
	Stocks string
	Budget float64
}

type MyJson struct {
	List struct {
		Meta struct {
			Count int    `json:"count"`
			Start int    `json:"start"`
			Type  string `json:"type"`
		} `json:"meta"`
		Resources []struct {
			Resource struct {
				Classname string `json:"classname"`
				Fields    struct {
					Name    string `json:"name"`
					Price   string `json:"price"`
					Symbol  string `json:"symbol"`
					Ts      string `json:"ts"`
					Type    string `json:"type"`
					Utctime string `json:"utctime"`
					Volume  string `json:"volume"`
				} `json:"fields"`
			} `json:"resource"`
		} `json:"resources"`
	} `json:"list"`
}

type Str string

func getFloat(s string) float64{
	i,err :=strconv.ParseFloat(s, 64)
	if err != nil {
        // handle error
        	fmt.Println(err)
        	os.Exit(2)
    	}
		
	truncated := float64(int(i * 100)) / 100
	
	return truncated
}

//function for calculating number of shares that can be bought
func BuyShares(price float64,money float64) (float64,int){

	i := (money)/price
	
	var c int
	c = int(i)
	
	left := money - (float64(c)* price)
	
	return left,c
	
}

//function to get price of company stocks from Yahoo finance Api
func getYahooPrice(sym []string) (MyJson,error){
	
	stocklist :=""

	for _, key:= range sym{
		stocklist += (key + ",")	
	}  

	stocklist = stocklist[:len(stocklist)-1]
	
	//fmt.Println(stocklist)

	urlstart := "http://finance.yahoo.com/webservice/v1/symbols/"
	urlend := "/quote?format=json"

	url2 := urlstart + stocklist + urlend

	var f MyJson
	
	res,err := http.Get(url2)
	if err!=nil{
		return f,err
	}
	defer res.Body.Close()
	
	body, err := ioutil.ReadAll(res.Body)
	if err!= nil{
		return f,err
	}
	
	

	err = json.Unmarshal(body,&f)
	if err!=nil{
		return f,err
	}

	//fmt.Println(f)
	
	return f,nil
}

func (t *Str) Compute(s Input, reply *CalcResp) error{

	var sym [10]string
	var percent [10]float64
	var price [10]float64
	var count [10]int
	var money [10]float64
	var balc [10]float64
	var err error
	var total float64
	var left float64
	var res string
	
               
		//splitting string request string to process request
		str := strings.Split(s.Stocks,",")
		l :=len(str)
		
		var k int

		for k=0;k<l;k++{
			s1 := strings.Split(str[k],":")
			sym[k] = s1[0]
			
			s2 :=strings.Split(s1[1],"%")
			percent[k],err =strconv.ParseFloat(s2[0],64)
			//fmt.Println(sym[k],"\t",percent[k])
			total = total + percent[k]	
		}
		if total!=100{
			return errors.New("Percentages not sum up to 100 ")		
		}	

	
	var f MyJson

	//calling method to get price
	f,err = getYahooPrice(sym[:])
	if err !=nil{
		 return errors.New("Error in getting price from YahooFinanceAPi")
	}

	m := len(f.List.Resources)
	if(m!=l){
		return errors.New("Error in stocks symbol")
	}
	/*for k=0;k<l;k++{
		fmt.Println(f.List.Resources[k].Resource.Fields.Price,"\t",f.List.Resources[k].Resource.Fields.Symbol)
	}*/

	for k=0;k<l;k++{
		price[k] = getFloat(f.List.Resources[k].Resource.Fields.Price)
		money[k] = (percent[k] *s.Budget)/100
		balc[k],count[k] = BuyShares(price[k],money[k])
		left = left + balc[k]
		res = res + sym[k] + ":" + strconv.Itoa(count[k]) + ":$" + fmt.Sprintf("%.2f",price[k])
		if k!=l-1{
			res = res + ","		
		}
		
	}
	Id++

	cr[Id].TradeId = Id+1
	cr[Id].StocksInfo = res
	cr[Id].Balance = getFloat(fmt.Sprintf("%.2f",left))
	
	*reply = cr[Id]
	return nil
}

func (t *Str) Portfolio(i int, reply *PortfolioResp) error{
	
	var x int

	if i<=0 || i-1>Id{
		return errors.New("Trade id not found")	
	}

	for x=0;x<Id;x++{
		if(cr[x].TradeId==i){
			break		
		}
	
	}
	//fmt.Println("checking portfolio : ",x,"   value in struct :",cr[x].TradeId)
	if x==Id && cr[x].TradeId!=i{
		return errors.New("Trade id not found")	
	}
	
	s := cr[x].StocksInfo
	data_res :=strings.Split(s,",")

	a:=len(data_res)

	
	var count [10]int
	var price [10]float64
	var sym [10]string
	var err error
	var k int

	for k=0;k<a;k++{
		s1 := strings.Split(data_res[k],":")
		sym[k] = s1[0]
		count[k],err = strconv.Atoi(s1[1])
		s2 :=strings.Split(s1[2],"$")
		price[k] = getFloat(s2[1])
		//fmt.Println(sym[k],"\t",count[k],"\t",price[k])	
	}

	var f MyJson
	
	f,err = getYahooPrice(sym[:])
	if err !=nil{
		 return errors.New("Error in getting price from YahooFinanceAPi")
	}

	var newprice [10]float64
	var amount [10]float64
	var sign [10]string
	var curval float64
	var str string
	
	for k=0;k<a;k++{
		newprice[k] = getFloat(f.List.Resources[k].Resource.Fields.Price)
		if newprice[k] < price[k]{
			sign[k]="-"	
		}else if newprice[k] > price[k]{
			sign[k]="+"	
		}else{
			sign[k]=""
		}
		amount[k] = newprice[k] * float64(count[k])
		curval = curval +amount[k]
		str = str + sym[k] + ":" + strconv.Itoa(count[k]) + ":" + sign[k] + "$" + fmt.Sprintf("%.2f",newprice[k])
		if k!=a-1{
			str = str + ","		
		}
	}
	

	var pr PortfolioResp
	
	pr.StocksInfo = str
	pr.CurMarketVal = getFloat(fmt.Sprintf("%.2f",curval))
	pr.Balance = cr[x].Balance
	
	*reply = pr
	return nil
}

func main(){
	Id = -1
	ed := new(Str)
	rpc.Register(ed)
	
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":1234")
   	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	    for {
		conn, err := listener.Accept()
		if err != nil {
		    continue
		}
		jsonrpc.ServeConn(conn)
	    }	
}

func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}
