package databuilder

type TestStruct1 struct {
}

type TestStruct2 struct {
}

func DBTestFunc(_ TestStruct1) (TestStruct2, error) {
	return TestStruct2{}, nil
}

func DBTestFunc2() (TestStruct1, error) {
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
