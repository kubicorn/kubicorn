package test

func init() {
	nilSlice := []string(nil)
	marshalCases = append(marshalCases,
		[]interface{}{"hello"},
		nilSlice,
		&nilSlice,
		selectedMarshalCase{[]byte{1,2,3}},
	)
	unmarshalCases = append(unmarshalCases, unmarshalCase{
		ptr:   (*[]string)(nil),
		input: "null",
	}, unmarshalCase{
		ptr:   (*[]string)(nil),
		input: "[]",
	}, unmarshalCase{
		ptr:   (*[]byte)(nil),
		input: "[1,2,3]",
	}, unmarshalCase{
		ptr:   (*[]byte)(nil),
		input: `"aGVsbG8="`,
		selected: true,
	})
}
