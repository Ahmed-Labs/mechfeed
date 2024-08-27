package filter

import "testing"

func TestFilter(t *testing.T) {
	t.Run("keywords match", func(t *testing.T) {
		const content = "HAHhAHAH omg WTB Kaze for free"
		var Keywords = "WTB,Kaze,-Red"
		got := FilterKeywords(content, Keywords)
		expect := true
	
		if got != expect {
			t.Errorf("got %t expect %t", got, expect)
		}
	})
	t.Run("keywords don't match", func(t *testing.T) {
		const content = "HAHhAHAH omg WTB a red Kaze for free"
		var Keywords = "WTB,Kaze,-Red"
		got := FilterKeywords(content, Keywords)
		expect := false
	
		if got != expect {
			t.Errorf("got %t expect %t", got, expect)
		}
	})
	t.Run("keywords don't match part of a word", func(t *testing.T) {
		const content = "HAHhAHAH omg WTbuy a kaze for free"
		var Keywords = "WTB,Kaze,-Red"
		got := FilterKeywords(content, Keywords)
		expect := false
	
		if got != expect {
			t.Errorf("got %t expect %t", got, expect)
		}
	})
	t.Run("empty keywords don't match", func(t *testing.T) {
		const content = "HAHhAHAH omg WTbuy a Kamikaze for free"
		var Keywords = ",   ,,"
		got := FilterKeywords(content, Keywords)
		expect := false
	
		if got != expect {
			t.Errorf("got %t expect %t", got, expect)
		}
	})
}
