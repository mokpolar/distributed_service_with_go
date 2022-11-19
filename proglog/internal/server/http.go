package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// 서버의 주소를 파라미터로 받아서 *http.Server를 리턴
// gorilla/mux 패키지를 사용하여 리퀘스트를 처리할 라우터를 생성
// / 엔드포인트를 호출하는 POST 요청은 produceHandler가 처리하여 레코드를 로그에 추가
// / 엔드포인트를 호출하는 GET 요청은 consumeHandler가 처리하여 로그에서 레코드를 읽음
// 생성한 httpServer는 *net/http.Server로 다시 래핑하여 ListenAndServer()를 이용해서 요청을 처리할 수 있음
func NewHTTPServer(addr string) *http.Server {
	httpsrv := newHTTPServer()
	r := mux.NewRouter()
	r.HandleFunc("/", httpsrv.handleProduce).Methods("POST")
	r.HandleFunc("/", httpsrv.handleConsume).Methods("GET")

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}

}

// 서버는 로그를 참조하고, 참조하는 로그를 핸들러에 전달한다.
// ProduceRequest는 호출자가 로그에 추가하길 원하는 레코드를 담고,
// ProduceResponse는 호출자에게 저장한 오프셋을 알려준다.
// ConsumeRequest는 호출자가 읽길 원하는 레코드의 오프셋을 담고,
// ConsumeResponse는 오프셋에 위치하는 레코드를 보내준다.

type httpServer struct {
	Log *Log // Log 구조체 포인터
}

func newHTTPServer() *httpServer { // *httpServer means that the function returns a pointer to an httpServer
	return &httpServer{
		Log: NewLog(), // Log 구조체 포인터를 생성
	}
}

type ProduceRequest struct {
	Record Record `json:"record"`
}

type ProduceResponse struct {
	Offset uint64 `json:"offset"`
}

type ConsumeRequest struct {
	Offset uint64 `json:"offset"`
}

type ConsumeResponse struct {
	Record Record `json:"record"`
}

// produce Handler의 3 단계 구현
// 요청을 구조체로 디코딩하고,
// 로그에 추가한 다음
// 오프셋을 구조체에 담아 인코딩하여 응답

func (s *httpServer) handleProduce(w http.ResponseWriter, r *http.Request) {

	// 요청을 구조체로 디코딩
	// 요청의 바디를 읽어서 ProduceRequest 구조체로 디코딩
	// 디코딩에 실패하면 400 에러를 반환
	// 디코딩에 성공하면 로그에 추가하고 오프셋을 구조체에 담아 인코딩하여 응답
	var req ProduceRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 로그에 추가
	// ProduceRequest 구조체의 Record 필드를 로그에 추가
	// 추가에 실패하면 500 에러를 반환
	// 추가에 성공하면 오프셋을 ProduceResponse 구조체에 담아 인코딩
	off, err := s.Log.Append(req.Record)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 오프셋을 구조체에 담아 인코딩
	// ProduceResponse 구조체를 인코딩
	// 인코딩에 실패하면 500 에러를 반환
	// 인코딩에 성공하면 응답
	res := ProduceResponse{Offset: off}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// consume 핸들러는 produce 핸들러와 비슷한 구조이지만 Read(offset uint64)를 호출하여
// 로그에서 레코드를 읽어낸다.
// 이 핸들러는 좀 더 많은 에러 체크를 하여 정확한 상태 코드를 클라이언트에 제공한다.
// 서버가 요청을 핸들링할 수 없다는 에러도 있고,
// 클라이언트가 요청한 레코드가 존재하지 않는다는 에러도 있다.
func (s *httpServer) handleConsume(w http.ResponseWriter, r *http.Request) {
	var req ConsumeRequest
	err := json.NewDecoder(r.Body).Decode(&req) // & means that the function returns a pointer to an httpServer
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	record, err := s.Log.Read(req.Offset)
	if err == ErrOffsetNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	res := ConsumeResponse{Record: record}
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
