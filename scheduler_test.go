package scheduler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	log "log"
)

var s *Driver

func TestMain(m *testing.M) {
	s = &Driver{
		Jobs: map[string]Job{
			"say_lala": {
				CronTime: "*/5 * * * * *",
				Do: func() {
					log.Println("lala")
				},
			},
		},
	}
	m.Run()
}
func TestHandleCmd(t *testing.T) {
	router := httprouter.New()
	router.GET("/runNow", func(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
		s.HandleCmd(rw, r)
	})
	{
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/runNow?job=say_lala", nil)
		router.ServeHTTP(recorder, request)
		assert.Equal(t, recorder.Code, http.StatusOK)
	}
	{
		recorder := httptest.NewRecorder()
		request, _ := http.NewRequest("GET", "/runNow?job=lalaland", nil)
		router.ServeHTTP(recorder, request)
		assert.Equal(t, recorder.Code, http.StatusNotFound)
	}
}
