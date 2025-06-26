package test

import (
	"context"
	"testing"
	"time"

	"github.com/crewcrew23/memo/internal/cache"
	"github.com/crewcrew23/memo/pkg/memo"
)

type TestData struct {
	Value int
}

func TestSetValue(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	c.Set("key", testData, time.Second*5)

	if _, err := c.Get("key"); err != nil {
		t.Fail()
	}

}

func TestSetValueWithContext(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}
	cancel()
}

func TestSetValueWithContext_Cancel(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	cancel()

	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err == nil {
		t.Fail()
	}

}

func TestGetValueWithContext(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	if _, err := c.GetWithContext(ctx, "key"); err != nil {
		t.Fail()
	}

	cancel()
}

func TestGetValueWithContext_Cancel(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	cancel()
	if _, err := c.GetWithContext(ctx, "key"); err == nil {
		t.Fail()
	}

}

func TestMarshal(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	if _, err := c.MarshalJSON(); err != nil {
		t.Fail()
	}

	cancel()
}

func TestMarshalWithContext(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	if _, err := c.MarshalJSONWithContext(ctx); err != nil {
		t.Fail()
	}
	cancel()
}

func TestMarshalWithContext_Cancel(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	cancel()
	if _, err := c.MarshalJSONWithContext(ctx); err == nil {
		t.Fail()
	}
}

func TestUnmarshal(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	bytes, _ := c.MarshalJSON()

	uc := memo.New[*TestData]()
	if err := uc.UnmarshalJSON(bytes); err != nil {
		t.Fail()
	}

	cancel()
}

func TestUnmarshalWithContext(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	bytes, _ := c.MarshalJSON()

	uc := memo.New[*TestData]()
	if err := uc.UnmarshalJSONWithContext(ctx, bytes); err != nil {
		t.Fail()
	}

	cancel()
}

func TestUnmarshalWithContext_Cancel(t *testing.T) {
	testData := &TestData{5}
	c := memo.New[*TestData]()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	if err := c.SetWithContext(ctx, "key", testData, time.Second*5); err != nil {
		t.Fail()
	}

	bytes, _ := c.MarshalJSON()

	uc := memo.New[*TestData]()

	cancel()
	if err := uc.UnmarshalJSONWithContext(ctx, bytes); err == nil {
		t.Fail()
	}
}

func TestCleaner(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	c := cache.New[*TestData](ctx, cancel)
	cache.StartClean(c, ctx, time.Millisecond*10)

	c.Set("key", &TestData{5}, time.Millisecond*1)

	time.Sleep(time.Millisecond * 100)

	if _, err := c.Get("key"); err == nil {
		t.Fail()
	}
}

func TestCloseConn(t *testing.T) {
	c := memo.New[*TestData]()
	c.Close()

	if _, err := c.Get("key"); err == nil {
		t.Fail()
	}
}
