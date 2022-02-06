package databuilder

type TestStruct1 struct {
}

type TestStruct2 struct {
}

type TestInter interface {
}

func DBTestFunc(_ TestStruct1) (TestStruct2, error) {
	return TestStruct2{}, nil
}

func DBTestFunc2() (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFunc3(_ TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid1(_ int) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid2(_ TestStruct2) TestStruct1 {
	return TestStruct1{}
}

func DBTestFuncInvalid3(...TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid4(_ *TestStruct2) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid5(_ TestStruct1) (TestStruct1, error) {
	return TestStruct1{}, nil
}

func DBTestFuncInvalid6(_ TestStruct2) (int, error) {
	return 0, nil
}

func DBTestFuncInvalid7(_ TestStruct1) (TestStruct1, int) {
	return TestStruct1{}, 0
}

func DBTestFuncInvalid8(_ TestStruct1) (TestStruct1, TestInter) {
	return TestStruct1{}, nil
}
