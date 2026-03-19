package generic

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/rcrowley/go-metrics"
)

var countersMap = cmap.New()

func ProcessMetrics(c *gin.Context) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	items := countersMap.Items()
	globalCounter := make(map[string]interface{})
	for k, v := range items {
		myCounter, ok := v.(metrics.Counter)
		if ok {
			globalCounter[k] = myCounter.Count()
		} else {
			globalCounter[k] = v
		}
	}

	data := map[string]interface{}{
		"systemMetrics": mem,
		"customMetrics": globalCounter,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error serializing metrics"})
		return
	}

	c.Data(http.StatusOK, "application/json", jsonData)
}

func IncrementCounter(keys ...string) {
	for i := 0; i < len(keys); i++ {
		var myCounter metrics.Counter
		objCounter, ok := countersMap.Get(keys[i])
		if !ok {
			newCounter := metrics.NewCounter()
			countersMap.Set(keys[i], newCounter)
			myCounter = newCounter
		} else {
			myCounter = objCounter.(metrics.Counter)
		}

		myCounter.Inc(1)
	}
}

const (
	TAG_COUNTER_AVAIL        = "AVAIL"
	TAG_COUNTER_AVAIL_OK     = "AVAIL_OK"
	TAG_COUNTER_AVAIL_ERRORS = "AVAIL_ERRORS"

	TAG_COUNTER_PREBOOK        = "PREBOOK"
	TAG_COUNTER_PREBOOK_OK     = "PREBOOK_OK"
	TAG_COUNTER_PREBOOK_ERRORS = "PREBOOK_ERRORS"

	TAG_COUNTER_BOOK        = "BOOK"
	TAG_COUNTER_BOOK_OK     = "BOOK_OK"
	TAG_COUNTER_BOOK_ERRORS = "BOOK_ERRORS"

	TAG_COUNTER_CANCEL        = "CANCEL"
	TAG_COUNTER_CANCEL_OK     = "CANCEL_OK"
	TAG_COUNTER_CANCEL_ERRORS = "CANCEL_ERRORS"

	TAG_COUNTER_ERRORS = "ERRORS"
	TAG_COUNTER_TOTAL  = "TOTAL"
)
