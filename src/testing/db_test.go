package testing

import (
	"fmt"
	"schoolbot/internal/model"
	"schoolbot/net/campusboard"
	"testing"

	"github.com/objectbox/objectbox-go/objectbox"
)

func initObjectBox() *objectbox.ObjectBox {
	objectBox, err := objectbox.NewBuilder().Model(model.ObjectBoxModel()).Build()
	if err != nil {
		panic(err)
	}
	return objectBox
}

func TestMain(t *testing.T) {
	fmt.Println("Beginning main Test")

	// load objectbox
	ob := initObjectBox()
	defer ob.Close()

	_ = model.BoxForClass(ob)

	fmt.Println("Main Test complete")
}

func TestDB(t *testing.T) {
	t.Log("Database test")

	// load objectbox
	ob := initObjectBox()
	defer ob.Close()

	box := model.BoxForClass(ob)

	t.Log("clearing DB")
	box.RemoveAll()
	t.Log("DB Cleared")

	campusboard.UpdateDB(box)

	clsList, _ := box.GetAll()

	for _, cls := range clsList {
		t.Logf("[%02d] %s %v", cls.Id, cls.Name, cls.Majors)
	}
}

func putExampleClass(box *model.ClassBox) (uint64, error) {
	// Create & put
	return box.Put(&model.Class{
		Name:   "Example Class",
		Abbr:   "ex",
		Majors: make([]string, 5, 5),
	})
}
