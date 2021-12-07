package requestStructure

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
	
	"github.com/gorilla/mux"
	"github.com/yomiji/gkBoot"
	"github.com/yomiji/gkBoot/request"
)

type ManualRequest struct {
	Name   string      `request:"header"`
	Count  int         `request:"header" alias:"Count-Var"`
	Scores []int8      `request:"query"`
	Tests  []complex64 `request:"query" json:"tests"`
}

func (m ManualRequest) Decode(ctx context.Context, httpRequest *http.Request) (request interface{}, err error) {
	name := httpRequest.Header.Get("Name")
	countVar := httpRequest.Header.Get("Count-Var")
	count, err := strconv.Atoi(countVar)
	if err != nil {
		return nil, err
	}
	scores := make([]int8, 0)
	scoresQuery := httpRequest.URL.Query().Get("scores")
	scoresVar := strings.Split(scoresQuery, ",")
	for _, score := range scoresVar {
		if sn, e := strconv.ParseInt(strings.TrimSpace(score), 10, 8); e == nil {
			scores = append(scores,int8(sn))
		}
	}
	
	tests := make([]complex64, 0)
	testsQuery := httpRequest.URL.Query().Get("tests")
	testsVar := strings.Split(testsQuery, ",")
	for _, test := range testsVar {
		if tn, e := strconv.ParseComplex(strings.TrimSpace(test), 64); e == nil {
			tests = append(tests, complex64(tn))
		}
	}
	
	return &ManualRequest{
		Name:   name,
		Count:  count,
		Scores: scores,
		Tests:  tests,
	}, nil
}

func (m ManualRequest) Info() request.HttpRouteInfo {
	panic("implement me")
}

func BenchmarkManualRequestDecoderMixTypes(b *testing.B) {
	mReq := new(ManualRequest)
	decoder := mReq.Decode
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Name", "testValue")
	request.Header.Set("Count-Var", "3")
	q := request.URL.Query()
	q.Set("Scores", "18,27")
	q.Set("tests", "13+4i, 1+3i")
	request.URL.RawQuery = q.Encode()
	b.ResetTimer()
	if _,err := decoder(context.TODO(), request); err != nil {
		b.FailNow()
		return
	}
	for i := 0; i < b.N; i++ {
		decoder(context.TODO(), request)
	}
}

type MixTypeRequest struct {
	Name   string      `request:"header"`
	Count  int         `request:"header" alias:"Count-Var"`
	Scores []int8      `request:"query"`
	Tests  []complex64 `request:"query" json:"tests"`
}

func (m MixTypeRequest) Info() request.HttpRouteInfo {
	panic("implement me")
}

func TestGenerateRequestDecoderMixType(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(MixTypeRequest{})
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Name", "testValue")
	request.Header.Set("Count-Var", "3")
	q := request.URL.Query()
	q.Set("Scores", "18,27")
	q.Set("tests", "13+4i, 1+3i")
	request.URL.RawQuery = q.Encode()
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*MixTypeRequest); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.Name != "testValue" || v.Count != 3 || len(v.Scores) != 2 || len(v.Tests) != 2 {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}

func BenchmarkGenerateRequestDecoderMixTypes(b *testing.B) {
	decoder, _ := gkBoot.GenerateRequestDecoder(MixTypeRequest{})
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Name", "testValue")
	request.Header.Set("Count-Var", "3")
	q := request.URL.Query()
	q.Set("Scores", "18,27")
	q.Set("tests", "13+4i, 1+3i")
	request.URL.RawQuery = q.Encode()
	b.ResetTimer()
	if _,err := decoder(context.TODO(), request); err != nil {
		b.FailNow()
		return
	}
	for i := 0; i < b.N; i++ {
		decoder(context.TODO(), request)
	}
}

func TestGenerateRequestDecoderMixedRequestDoesNotDupe(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(MixTypeRequest))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Name", "testValue")
	request.Header.Set("Count-Var", "3")
	q := request.URL.Query()
	q.Set("Scores", "18,27")
	q.Set("tests", "13+4i, 1+3i")
	request.URL.RawQuery = q.Encode()
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		request2, _ := http.NewRequest("GET", "http://localhost", nil)
		request2.Header.Set("Name", "testValue")
		request2.Header.Set("Count-Var", "4")
		q := request2.URL.Query()
		q.Set("tests", "13+4i, 1+3i")
		request2.URL.RawQuery = q.Encode()
		val, err = decoder(context.TODO(), request2)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*MixTypeRequest); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.Name != "testValue" || v.Count != 4 || len(v.Scores) == 2 || len(v.Tests) != 2 {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}

type EmbeddedRequest struct {
	Name string `request:"header"`
	EmbeddableRequest1
}

type EmbeddableRequest1 struct {
	Count complex128 `request:"header" alias:"Count-Flag"`
	*NestedRequest2
}

func (e EmbeddableRequest1) Info() request.HttpRouteInfo {
	panic("implement me")
}

type NestedRequest2 struct {
	Stat int32 `request:"query" json:"stat"`
}

func TestGenerateRequestDecoderEmbedded(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(EmbeddedRequest))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Name", "testValue")
	request.Header.Set("Count-Flag", "3")
	q := request.URL.Query()
	q.Set("stat", "1345")
	request.URL.RawQuery = q.Encode()
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*EmbeddedRequest); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.Name != "testValue" || v.Count != 3 || v.Stat != 1345 {
			t.Fatalf("values do not match: %+v, stat: %+v", v, *v.NestedRequest2)
		}
	}
}

type RegularObject struct {
	RequestHeader string `request:"header" alias:"Request-Header" json:"requestHeader"`
	RegularEmbed
}

type RegularEmbed struct {
	NormalInt   int    `json:"normalInt"`
	TypicalJson string `json:"typicalJson"`
}

func (r RegularEmbed) Info() request.HttpRouteInfo {
	panic("implement me")
}

func TestGenerateRequestDecoderWorksWithRegularEmbed(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(RegularObject))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Request-Header", "testValue")
	request.Header.Set("normalInt", "3")
	q := request.URL.Query()
	q.Set("normalInt", "1345")
	request.URL.RawQuery = q.Encode()
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*RegularObject); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.RequestHeader != "testValue" || v.NormalInt == 3 || v.TypicalJson != "" {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}

type PointedObjectTest struct {
	RequestHeader string `request:"header" alias:"Request-Header" json:"requestHeader"`
	*referenceEmbed
}

type referenceEmbed struct {
	NormalInt   int    `json:"normalInt"`
	TypicalJson string `json:"typicalJson"`
}

func (r referenceEmbed) Info() request.HttpRouteInfo {
	panic("implement me")
}

func TestGenerateRequestDecoderWorksWithUnexportedReferenceEmbed(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(PointedObjectTest))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Request-Header", "testValue")
	request.Header.Set("normalInt", "3")
	q := request.URL.Query()
	q.Set("normalInt", "1345")
	request.URL.RawQuery = q.Encode()
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*PointedObjectTest); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.RequestHeader != "testValue" || v.referenceEmbed != nil {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}

type FormTestObject struct {
	Header string `request:"header"`
	Body Form `request:"form" json:"body"`
}

func (f FormTestObject) Info() request.HttpRouteInfo {
	panic("implement me")
}

type Form struct {
	Value string `json:"someValue"`
	Count int `json:"count"`
}

func TestGenerateRequestDecoderForm(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(FormTestObject))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Header", "testValue")
	request.Body = io.NopCloser(strings.NewReader("{\"someValue\":\"val\",\"count\":44}"))
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*FormTestObject); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.Header != "testValue" || v.Body.Count != 44 || v.Body.Value != "val" {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}

func BenchmarkGenerateRequestDecoderForm(b *testing.B) {
	decoder, _ := gkBoot.GenerateRequestDecoder(new(FormTestObject))

	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("Header", "testValue")
	request.Body = io.NopCloser(strings.NewReader("{\"someValue\":\"val\",\"count\":44}"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decoder(context.TODO(), request)
	}
}

type RequiredField struct {
	TheRightHeader bool `request:"header!"`
}

func (r RequiredField) Info() request.HttpRouteInfo {
	panic("implement me")
}

func TestGenerateRequestDecoderRequiredFieldsNotFound(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(RequiredField))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("NotTheRightHeader", "testValue")
	if decoder == nil {
		t.Fail()
	} else {
		_, err := decoder(context.TODO(), request)
		if err == nil {
			t.Fatal("did not receive expected error")
		}
	}
}

func TestGenerateRequestDecoderRequiredFieldsFound(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(RequiredField))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("TheRightHeader", "true")
	if decoder == nil {
		t.Fail()
	} else {
		_, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("expected to succeed but got: %s", err.Error())
		}
	}
}

type PointerTest struct {
	P1int *int `request:"header" json:"p1"`
	P2string *string `request:"header" json:"p2"`
	P3float32Slice []**float32 `request:"header" json:"p3"`
	P4Blank **int `request:"header" json:"p4"`
}

func (p PointerTest) Info() request.HttpRouteInfo {
	panic("implement me")
}

func TestGenerateRequestDecoderHandlesPointers(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(PointerTest))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost", nil)
	request.Header.Set("p1", "123")
	request.Header.Set("p2", "test")
	request.Header.Set("p3", "12.3, , 34.5")
	request.Header.Set("p4", "")
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*PointerTest); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if **v.P4Blank != 0 || *v.P1int != 123 || *v.P2string != "test" || (len(v.P3float32Slice) != 3 || **v.P3float32Slice[0] != 12.3) {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}

type PathTest struct {
	Name string `request:"path" json:"name"`
	Count int `request:"path" json:"count"`
}

func (p PathTest) Info() request.HttpRouteInfo {
	panic("implement me")
}

func TestGenerateRequestDecoderHandlesPathVars(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(PathTest))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost/name/count", nil)
	request = mux.SetURLVars(request, map[string]string{"name":"billy", "count":"121"})
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*PathTest); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.Count != 121 || v.Name != "billy" {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}

type JsonDecoderTest struct {
	Name string `json:"name"`
	Count int `json:"count"`
	Test float32 `json:"test"`
	gkBoot.JSONBody
}

func (j JsonDecoderTest) Info() request.HttpRouteInfo {
	panic("implement me")
}

func TestGenerateJSONBodyDecoder(t *testing.T) {
	decoder, err := gkBoot.GenerateRequestDecoder(new(JsonDecoderTest))
	if err != nil {
		t.Fail()
	}
	request, _ := http.NewRequest("GET", "http://localhost/name/count", nil)
	request.Body = io.NopCloser(strings.NewReader("{\"name\":\"val\",\"count\":44,\"test\":21.1}"))
	if decoder == nil {
		t.Fail()
	} else {
		val, err := decoder(context.TODO(), request)
		if err != nil {
			t.Fatalf("basic decoder failure: %s", err.Error())
		}
		if v, ok := val.(*JsonDecoderTest); !ok {
			t.Fatalf("type not correct: %T", val)
		} else if v.Count != 44 || v.Name != "val" || v.Test != 21.1 {
			t.Fatalf("values do not match: %+v", v)
		}
	}
}
