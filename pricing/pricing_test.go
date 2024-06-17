package pricing

import (
	"context"
	// "os"
	"sync"
	"testing"
	"time"

	// "github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const (
	PROCESS_INTERVAL = 5 * time.Second
	MAX_PRICE_DELAY  = time.Minute
)

func TestMedian(t *testing.T) {
	logrus.Warn(".......TestMedian")
	pf := NewMarketPriceService("", 7, nil, nil)
	values := []float64{1.0, 2.0, 3.0, 0.5, -9.0, -8.0, 2.0}
	res := pf.median(values)
	if res != 1.0 {
		t.Fatalf("Expected median of 1.0 but got %v", res)
	}
	values = []float64{-9.0, -8.0, 0.5, 1.0, 2.0, 2.0, 3.0}
	res = pf.median(values)
	if res != 1.0 {
		t.Fatalf("Expected median of 1.0 but got %v", res)
	}
	values = []float64{-9.0, -8.0, 0.5, 1.0, 2.0, 2.0}
	res = pf.median(values)
	if res != 0.75 {
		t.Fatalf("Expected median of 0.75 but got %v", res)
	}
}

func TestCloseEnough(t *testing.T) {
	logrus.Warn(".......TestCloseEnough")
	value := 1.0
	target := 0.99
	tolerance := 0.1
	if !closeEnough(value, target, tolerance) {
		t.Fatalf("Expected value %v to be within tolerance %v of target %v", value, tolerance, target)
	}
	value = 123.0
	target = 124.0
	tolerance = 0.01
	if !closeEnough(value, target, tolerance) {
		t.Fatalf("Expected value %v to be within tolerance %v of target %v", value, tolerance, target)
	}
	value = 124.0
	target = 123.0
	tolerance = 0.01
	if !closeEnough(value, target, tolerance) {
		t.Fatalf("Expected value %v to be within tolerance %v of target %v", value, tolerance, target)
	}
	value = 123.0
	target = 124.0
	tolerance = 0.008
	if closeEnough(value, target, tolerance) {
		t.Fatalf("Did not expect value %v to be within tolerance %v of target %v", value, tolerance, target)
	}
	value = 124.0
	target = 123.0
	tolerance = 0.008
	if closeEnough(value, target, tolerance) {
		t.Fatalf("Did not expect value %v to be within tolerance %v of target %v", value, tolerance, target)
	}
}

func TestConsistencyCheck(t *testing.T) {
	logrus.Warn(".......TestConsistencyCheck")
	pf := NewMarketPriceService("", 7, nil, nil)
	values := []float64{1.0, 2.0, 3.0, 0.5, -9.0, -8.0, 2.0}
	median1, consistent := pf.checkDataConsistency(values)
	if consistent {
		t.Fatalf("Should be reported as inconsistent, but found to be consistent, median is %v", median1)
	}
	values = []float64{1.0, 2.0, 3.0, 0.5, -9.0, -8.0, 2.0, 1.04}
	median1, consistent = pf.checkDataConsistency(values)
	if !consistent {
		t.Fatalf("Should be reported as consistent, but found to be inconsistent, median is %v", median1)
	}
}

type DummyAggregator struct {
	values [][]float64
	ch     chan []float64
}

func (da *DummyAggregator) start(ctx context.Context, wg *sync.WaitGroup) {
	go da.streamPrices(ctx)
	time.Sleep(100 * time.Millisecond)
	wg.Done()
}

func (da *DummyAggregator) streamPrices(ctx context.Context) {
	for _, value := range da.values {
		da.ch <- value
	}
}

func NewDummyAggregator(values [][]float64) *DummyAggregator {
	return &DummyAggregator{
		values: values,
		ch:     make(chan []float64),
	}
}

type PriceChecker struct {
	t        *testing.T
	expected []float64
	current  int64
}

func NewPriceChecker(t *testing.T, expected []float64) *PriceChecker {
	return &PriceChecker{
		t:        t,
		expected: expected,
		current:  0,
	}
}

func (pc *PriceChecker) UpdateIndexPrice(ctx context.Context, market_id string, index_price float64) error {
	if index_price != pc.expected[pc.current] {
		pc.t.Errorf("Index %d expected price %v but got %v", pc.current, pc.expected[pc.current], index_price)
	}
	pc.current++
	return nil
}

func TestMarketPriceJumpHandling(t *testing.T) {
	logrus.Warn(".......TestMarketPriceJumpHandling")
	values := [][]float64{{1345.6}, {1345.7}, {1098.3}, {918.2}, {892.3},
		{899.4}, {856.3}, {844.7}, {824.5}, {790.2}, {766.8},
		{759.4}, {761.4}, {757.8}, {604.3}, {699.1}, {878.4},
		{987.2}, {1099.2}, {1094.7}, {1095.8}, {1099.9}, {1107.0},
		{1110.1}, {1165.2}, {1187.3}, {1199.4}, {1234.5}, {1345.6},
		{1345.7}, {1098.3}, {1098.2}, {1098.7}, {1098.4}, {1089.6},
		{1089.3}, {1089.2}, {1089.1}, {1089.0}, {1088.9}, {1088.8},
		{1098.8}, {783.4}, {1095.0}}
	expected := []float64{1345.6, 1345.7, 757.8, 699.1, 1234.5, 
		1345.6, 1345.7, 1098.8, 1095.0}
	testMarketPriceWithDummyData(t, values, expected)
}

func TestMarketPriceTwoSourcesJumpHandling(t *testing.T) {
	values := [][]float64{{1345.6, 1344.6}, {1345.7, 1346.7}, {1098.3, 1097.3},
		{1098.2, 1099.2}, {1098.7, 1094.7}, {1098.4, 1095.3}, {1089.6, 1088.5},
		{1089.3, 1090.3}, {1089.2, 1087.5}, {1089.1, 1088.2}, {1089.0, 1089.6},
		{1088.9, 1088.9}, {1088.8, 1087.8}, {1098.8, 1095.8}, {1099.2, 195.8}}
	expected := []float64{1345.1, 1346.2, 1097.3, 1099.2}
	logrus.Warn("......TestMarketPriceTwoSourcesJumpHandling")
	testMarketPriceWithDummyData(t, values, expected)
}

func TestMarketPriceDivergentSourceJumpHandling(t *testing.T) {
	values := [][]float64{{1345.6, 1344.6, 1544.6}, {1345.7, 1346.7, 1546.7},
		{1098.3, 1097.3, 1297.3}, {1098.2, 1099.2, 1299.2},
		{1098.7, 1094.7, 1294.7}, {1098.4, 1095.3, 1295.3},
		{1089.6, 1088.5, 1288.5}, {1089.3, 1090.3, 1290.3},
		{1089.2, 1087.5, 1287.5}, {1089.1, 1088.2, 1288.2},
		{1089.0, 1089.6, 1289.6}, {1088.9, 1088.9, 1288.9},
		{1088.8, 1087.8, 1287.8}, {1098.8, 1095.8, 1295.8}}
	expected := []float64{1345.1, 1346.2, 1097.3}
	logrus.Warn(".......TestMarketPriceDivergentSourceJumpHandling")
	testMarketPriceWithDummyData(t, values, expected)
}

func testMarketPriceWithDummyData(t *testing.T, values [][]float64,
	expected []float64) {
	ctx := context.Background()
	ag := NewDummyAggregator(values)
	checker := NewPriceChecker(t, expected)
	mps := NewMarketPriceService("TEST", len(values[0]), ag.ch, checker)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		mps.start(ctx)
		ag.start(ctx, &wg)
	}()
	wg.Wait()
}

// Not really a test, code to get CMC coin listings...
//
// type ListingsResponse struct {
// 	Status struct {
// 		Timestamp    string `json:"timestamp"`
// 		ErrorCode    int    `json:"error_code"`
// 		ErrorMessage string `json:"error_message"`
// 	} `json:"status"`
// 	Data []struct {
// 		ID   int    `json:"id"`
// 		Name string `json:"name"`
// 	} `json:"data"`
// }
//
// func TestCMCCoins(t *testing.T) {

// 	godotenv.Load()
// 	apiKey := os.Getenv("COINMARKETCAP_API_KEY")
// 	// listingsURL := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/listings/latest"
// 	listingsURL := "https://pro-api.coinmarketcap.com/v1/cryptocurrency/map"
// 	client := &http.Client{}

// 	req, err := http.NewRequest("GET", listingsURL, nil)
// 	if err != nil {
// 		fmt.Println("Error creating request:", err)
// 		return
// 	}

// 	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		fmt.Println("Error executing request:", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		fmt.Println("Error reading response:", err)
// 		return
// 	}

// 	var listings ListingsResponse
// 	err = json.Unmarshal(body, &listings)
// 	if err != nil {
// 		fmt.Println("Error unmarshalling response:", err)
// 		return
// 	}

// 	for _, coin := range listings.Data {
// 		logrus.Printf("%v", coin)
// 	}
// }
