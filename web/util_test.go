package web

import (
	"net/url"
	"reflect"
	"testing"

	"celus-ti.com.br/qualitec/server/database/model"
)

func TestNewUrlQuery(t *testing.T) {
	testData := []struct {
		inputPath       string
		inputValues     url.Values
		clientsExpected []int16
		sitesExpected   []model.SiteID
		encoded         string
	}{
		{"/test",
			url.Values{
				"c": []string{"1"}, // Parâmetro especial para clientes
				"s": []string{"7"}, // Parâmetro especial para sites
				"v": []string{"value1"},
			},
			[]int16{1},        // Clients
			[]model.SiteID{7}, // Sites
			"/test?c=1&s=7&v=value1",
		},
		{"/test2",
			url.Values{
				"c": []string{"1", "2"}, // Parâmetro especial para clientes
				"s": []string{"7", "8"}, // Parâmetro especial para sites
			},
			[]int16{1, 2},        // Clients
			[]model.SiteID{7, 8}, // Sites
			"/test2?c=1&c=2&s=7&s=8",
		},
		{"/test3",
			url.Values{},
			[]int16{},        // Clients
			[]model.SiteID{}, // Sites
			"/test3",
		},
	}

	for _, data := range testData {
		u := NewURLQuery(data.inputValues, data.inputPath)

		// Test Sites
		if (len(data.sitesExpected) > 0 || len(u.Sites) > 0) && !reflect.DeepEqual(u.Sites, data.sitesExpected) {
			t.Errorf("Expected %v, got %v", data.sitesExpected, u.Sites)
		}

		// Test Clients
		if (len(data.clientsExpected) > 0 || len(u.Clients) > 0) && !reflect.DeepEqual(u.Clients, data.clientsExpected) {
			t.Errorf("Expected %v, got %v", data.clientsExpected, u.Clients)
		}

		encoded := u.Encode()
		if encoded != data.encoded {
			t.Errorf("Expected %s, got %s", data.encoded, encoded)
		}
	}

	data := testData[0]
	u := NewURLQuery(data.inputValues, data.inputPath)

	// Test new value
	got := u.GenURL("z", 10)
	expect := "/test?c=1&s=7&v=value1&z=10"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Testa se os valores internos não foram modificados
	got = u.Encode()
	if got != data.encoded {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Modifica um valor
	got = u.GenURL("v", "modificado")
	expect = "/test?c=1&s=7&v=modificado"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Testa se os valores internos não foram modificados
	got = u.Encode()
	if got != data.encoded {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Remove um valor
	got = u.GenURL("v", "")
	expect = "/test?c=1&s=7"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Testa a geração com dois parâmetros
	got = u.GenURLMulti2("t1", 1, "t2", 2)
	expect = "/test?c=1&s=7&t1=1&t2=2&v=value1"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Testa a geração com dois parâmetros sendo um exclusão
	got = u.GenURLMulti2("t1", 1, "v", "")
	expect = "/test?c=1&s=7&t1=1"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Testa se os valores internos não foram modificados
	got = u.Encode()
	if got != data.encoded {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	// Testa geração de URL BASE
	got = u.GenURLBase("/newurl")
	expect = "/newurl?c=1&s=7&v=value1"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	got = u.GenURLBase("/")
	expect = "/?c=1&s=7&v=value1"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	got = u.GenURLBaseMulti1("/", "t1", 1)
	expect = "/?s=1&t1=1"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	got = u.GenURLBaseMulti2("/", "t1", 1, "t2", 2)
	expect = "/?s=1&t1=1&t2=2"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	if u.HasSite(1) {
		t.Errorf("Expected false, got true")
	}

	if !u.HasSite(7) {
		t.Errorf("Expected true, got false")
	}

	// Testa geraão de url BASE vazia
	data = testData[2]
	u = NewURLQuery(data.inputValues, data.inputPath)
	got = u.GenURLBase("/newurl")
	expect = "/newurl"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

	got = u.GenURLBase("/")
	expect = "/"
	if got != expect {
		t.Errorf("Expected %s, got %s", expect, got)
	}

}
