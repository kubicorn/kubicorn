package ptrconvenient

func Int32Ptr(i int) *int32 {
	i32 := int32(i)
	return &i32
}


func Int64Ptr(i int) *int64 {
	i64 := int64(i)
	return &i64
}