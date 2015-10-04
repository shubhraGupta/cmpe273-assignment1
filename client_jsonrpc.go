package main

import("fmt"
	"os"
	"net/rpc/jsonrpc"
	"log"
	"strconv"
	
)

type CalcResp struct {
	TradeId int
	StocksInfo string
	Balance float64
}

type PortfolioResp struct {
	StocksInfo string
	CurMarketVal float64
	Balance float64
}

type Input struct{
	Stocks string
	Budget float64
}

func main(){
	if len(os.Args) <2{
		fmt.Println("input request")
        	log.Fatal(1)	
	}

	client, err := jsonrpc.Dial("tcp", "127.0.0.1:1234")

	if err != nil {
		log.Fatal("dialing:", err)
	}

	//For buying stocks request
	if len(os.Args) ==3{

		var s Input
		var calcResp CalcResp
		
		i, err := strconv.ParseFloat(os.Args[2], 64)

		s.Stocks =os.Args[1]
		s.Budget = i

		    err = client.Call("Str.Compute", s, &calcResp)
		    if err != nil {
			log.Fatal("error:", err)
		    }
	
		 fmt.Println("Trade id : ",calcResp.TradeId)
		 fmt.Println("Stocks purchased :",calcResp.StocksInfo)
		 fmt.Println("unvested amount : ",calcResp.Balance)
		 fmt.Println("")
	}

	//For portfolio request
	if len(os.Args) ==2{
		var pr PortfolioResp

		i,err := strconv.Atoi(os.Args[1])
		//var a int

		 err = client.Call("Str.Portfolio", i, &pr)
		    if err != nil {
			log.Fatal("error:", err)
		    }
		
		fmt.Println("Stocks info : ",pr.StocksInfo)
		 fmt.Println("Current Market Value :",pr.CurMarketVal)
		 fmt.Println("unvested amount : ",pr.Balance)
		 fmt.Println("")
	}

	
}
