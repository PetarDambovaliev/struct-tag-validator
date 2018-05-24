package validator

import (
	"testing"
	"os"
	"path/filepath"
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"strconv"
	"errors"
)

const cnt  = 0xC350 //50k

var modelsPath = "github.com/petar-dambovaliev/struct-tag-validator/models"

var declrTmp = `package models

import (
	"github.com/gobuffalo/uuid"
	"time"
)`

var structTmp = `

type %v struct {
	ID        uuid.UUID`+
	"`" +
	`json:"id" db:"id"` +
	"`\n" +
	`CreatedAt time.Time ` +
	"`" +
	`json:"created_at" db:"%v"` +
	"`\n" + `UpdatedAt time.Time ` +
	"`" +
	`json:"updated_at" db:"%v"` +
	"`\n}" +
	"\n   " +
	"\n" +
	"func testFunc(){}"

type structTpl struct {
	structName string
	createdAt string
	updatedAt string
	duplicateField string
}


func createModel(fileName string, structs []structTpl) {
	os.Mkdir("./models", 0755)

	var tmp = declrTmp

	for _, structTp := range structs {
		tmp = strings.Join([]string{
			tmp,
			strings.Join(
				[]string{
					fmt.Sprintf(
						structTmp,
						structTp.structName,
						structTp.createdAt,
						structTp.updatedAt,
					)},
				"\n"),
		}, "")
	}

	f, _ := os.Create(filepath.Join("models", fileName))
	f.WriteString(tmp)
	f.Close()
}

func Test_testValidate(t *testing.T) {
	r := require.New(t)

	structs := []structTpl{
		{
			"Customer",
			"created_at",
			"updated_at",
			"",
		},
		{
			"Customer1",
			"created_at",
			"updated_at",
			"",
		},
	}

	createModel("customer.go", structs)
	defer os.RemoveAll("./models")

	m := NewValidator(modelsPath)

	m.AddDefaultProcessors("db", "newtag")

	errs := m.Run()
	r.Empty(errs)
}

func Test_testValidateCustomProcessor(t *testing.T) {
	r := require.New(t)

	structs := []structTpl{
		{
			"Customer",
			"created_at",
			"updated_at",
			"",
		},
		{
			"Customer1",
			"created_at",
			"updated_at",
			"",
		},
	}

	createModel("customer.go", structs)
	defer os.RemoveAll("./models")

	m := NewValidator(modelsPath)

	m.AddProcessor("db", func(tag *Tag) []error {
		errs := []error{}

		if len(tag.value) > 2 {
			errs = append(errs, errors.New("Too long"))
		}

		return errs
	})

	errs := m.Run()

	r.Equal(4, len(errs))
}

func Test_testValidateDuplicates(t *testing.T) {
	r := require.New(t)
	structs := []structTpl{
		{
			"Customer",
			"created_at",
			"created_at",
			"created_at",
		},
		{
			"Customer1",
			"created_at",
			"updated_at",
			"updated_at",
		},
		{
			"Customer2",
			"created_at",
			"created_at",
			"created_at",
		},
	}

	createModel("customer.go", structs)
	createModel("customer1.go", structs)
	defer os.RemoveAll("./models")

	m := NewValidator(modelsPath)
	m.AddDefaultProcessors("db")

	errs := m.Run("Customer")

	r.Equal(1, len(m.packages))

	for _, pkg := range m.packages {
		r.Equal(1, len(pkg.Files))
		for fileName := range pkg.Files {
			r.Equal(true, strings.HasSuffix(fileName, "customer.go"))
		}
	}

	r.Len(errs, 2)
}

func Test_testValidateAllowDuplicates(t *testing.T) {
	r := require.New(t)
	structs := []structTpl{
		{
			"Customer",
			"created_at",
			"created_at",
			"created_at",
		},
		{
			"Customer1",
			"created_at",
			"updated_at",
			"updated_at",
		},
		{
			"Customer2",
			"created_at",
			"created_at",
			"created_at",
		},
	}

	createModel("customer.go", structs)
	createModel("customer1.go", structs)
	defer os.RemoveAll("./models")

	m := NewValidator(modelsPath)
	m.SetAllowDuplicates(true)
	m.AddDefaultProcessors("db")

	errs := m.Run("Customer")

	r.Len(errs, 0)
}

func Test_testValidator_ErrorsCount(t *testing.T)  {
	r := require.New(t)


	for i := 0; i < 44 ; i++ {
		structs := []structTpl{{
			"Customer" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
		},
		}

		createModel("Customer" + strconv.Itoa(i) + ".go", structs)
	}

	m := NewValidator(modelsPath)
	m.AddDefaultProcessors("db")
	errs := m.Run()

	r.Len(errs,44)

	os.RemoveAll("./models")
}

func BenchmarkModel_ValidateNoErrors(b *testing.B) {

	//We don't want to add the struct creation time into the benchmark
	//so we stop the timer
	b.StopTimer()

	//Let's stress the program and create 50k models
	for i := 0; i < cnt; i++ {
		structs := []structTpl{{
			"Customer" + strconv.Itoa(i),
			"created_at" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
			},
		}

		createModel("Customer" + strconv.Itoa(i) + ".go", structs)
	}

	b.StartTimer()

	for i := 0; i < 5; i++ {
		//Lets time the meat and potatoes of the benchmark
		b.Run("subtest", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m := NewValidator(modelsPath)
				m.AddDefaultProcessors("db")
				m.Run()
			}
		})
	}

	//Don't want to time the deletion of the files
	b.StopTimer()
	os.RemoveAll("./models")
}

func BenchmarkModel_ValidateWithErrors(b *testing.B) {

	//We don't want to add the struct creation time into the benchmark
	//so we stop the timer
	b.StopTimer()


	//Let's stress the program and create 50k models
	for i := 0; i < cnt ; i++ {
		structs := []structTpl{{
			"Customer" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
			"updated_at" + strconv.Itoa(i),
			},
		}

		createModel("Customer" + strconv.Itoa(i) + ".go", structs)
	}


	b.StartTimer()

	for i := 0; i < 5; i++ {
		//Lets time the meat and potatoes of the benchmark
		b.Run("subtest", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m := NewValidator(modelsPath)
				m.AddDefaultProcessors("db")
				m.Run()
			}
		})
	}

	//Don't want to time the deletion of the files
	b.StopTimer()
	os.RemoveAll("./models")
}
