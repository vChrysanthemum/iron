package iron

import "log"

func AssertErrIsNil1(ignore interface{}, err error) {
	AssertErrIsNil(err)
}

func AssertNotNil(obj interface{}) {
	if obj == nil {
		log.Panic("obj is nil")
	}
}

func AssertErrIsNil(err error) {
	if err != nil {
		log.Panic(err.Error())
	}
}

func AssertTrue(res bool) {
	if res == false {
		panic("AssertTrue invalid")
	}
}

func Ignore(r interface{}) {
}
