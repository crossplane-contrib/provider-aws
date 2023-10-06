/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	suffixTestNameMapper = "_test_namemapper"
	nameFiltered1        = "filtered1"
	nameFiltered2        = "filtered2"
	nameUnspecified      = "unspecified"
	nameMapperTest       = "mapper_test"
	suffixID             = "ID"
	suffixId             = "Id" //nolint:golint
	nameWith             = "nameWith"
	nameWithID           = nameWith + suffixID
	nameWithId           = nameWith + suffixId //nolint:golint
	prefixHTTP           = "HTTP"
	prefixHttp           = "Http" //nolint:golint
	nameVersion          = "Version"
	nameHTTPVersion      = prefixHTTP + nameVersion
	nameHttpVersion      = prefixHttp + nameVersion //nolint:golint
)

var (
	testNameMapper = NameMapper(func(s string) string {
		return s + suffixTestNameMapper
	})

	mapFilter = map[string]bool{
		nameFiltered1: true,
		nameFiltered2: true,
	}

	mapReplacerHTTPVersion = map[string]string{
		nameHTTPVersion: nameHttpVersion,
	}

	testNameFilter = nameFilter(func(s string) bool {
		return mapFilter[s]
	})
)

func TestApply(t *testing.T) {
	type fields struct {
		nameMappers mapperArr
		nameFilters filterArr
	}

	type args struct {
		opt []LateInitOption
	}

	tests := map[string]struct {
		fields fields
		args   args
	}{
		"EmptyOptions": {},
		"NonEmptyOptions": {
			fields: fields{
				nameMappers: []NameMapper{testNameMapper},
				nameFilters: []nameFilter{testNameFilter},
			},
			args: args{
				opt: []LateInitOption{testNameFilter, testNameMapper},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			want := &lateInitOptions{
				nameMappers: tt.fields.nameMappers,
				nameFilters: tt.fields.nameFilters,
			}

			got := lateInitOptions{}

			got.apply(tt.args.opt...)

			if diff := cmp.Diff(*want, got, cmp.AllowUnexported(lateInitOptions{}),
				cmp.Comparer(func(nm1, nm2 NameMapper) bool {
					return reflect.ValueOf(nm1).Pointer() == reflect.ValueOf(nm2).Pointer()
				}),
				cmp.Comparer(func(nf1, nf2 nameFilter) bool {
					return reflect.ValueOf(nf1).Pointer() == reflect.ValueOf(nf2).Pointer()
				})); diff != "" {
				t.Errorf("remove: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	type args struct {
		name string
	}

	tests := map[string]struct {
		fArr filterArr
		args args
		want bool
	}{
		"TestNoFilter": {
			fArr: nil,
			args: args{nameFiltered1},
			want: false,
		},
		"TestEmptyFilter": {
			fArr: filterArr{nil},
			args: args{nameFiltered1},
			want: false,
		},
		"TestFilteredName": {
			fArr: filterArr{testNameFilter},
			args: args{nameFiltered1},
			want: true,
		},
		"TestUnfilteredName": {
			fArr: filterArr{testNameFilter},
			args: args{nameUnspecified},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.fArr.filter(tt.args.name); got != tt.want {
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetName(t *testing.T) {
	type args struct {
		name string
	}

	tests := map[string]struct {
		mArr mapperArr
		args args
		want string
	}{
		"TestNoMapper": {
			mArr: nil,
			args: args{nameMapperTest},
			want: nameMapperTest,
		},
		"TestEmptyMapper": {
			mArr: mapperArr{nil},
			args: args{nameMapperTest},
			want: nameMapperTest,
		},
		"TestMappedName": {
			mArr: mapperArr{testNameMapper},
			args: args{nameMapperTest},
			want: nameMapperTest + suffixTestNameMapper,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tt.mArr.getName(tt.args.name); got != tt.want {
				t.Errorf("filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCanonicalNameFilter(t *testing.T) {
	type args struct {
		cNames []string
		name   string
	}

	tests := map[string]struct {
		args args
		want bool
	}{
		"EmptyCNameFilter": {
			args: args{
				name: nameFiltered1,
			},
			want: false,
		},
		"NonEmptyCNameFilter1": {
			args: args{
				cNames: []string{
					nameFiltered1,
					nameFiltered2,
				},
				name: nameFiltered1,
			},
			want: true,
		},
		"NonEmptyCNameFilter2": {
			args: args{
				cNames: []string{
					nameFiltered1,
					nameFiltered2,
				},
				name: nameFiltered2,
			},
			want: true,
		},
		"NonEmptyCNameUnfiltered": {
			args: args{
				cNames: []string{
					nameFiltered1,
					nameFiltered2,
				},
				name: nameUnspecified,
			},
			want: false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := canonicalNameFilter(tt.args.cNames...).filter(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("canonicalNameFilter().filter(...) = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSuffixReplacer(t *testing.T) {
	type args struct {
		suffix  string
		replace string
		name    string
	}
	tests := map[string]struct {
		args args
		want string
	}{
		"TestWithSuffix": {
			args: args{
				suffix:  suffixID,
				replace: suffixId,
				name:    nameWithID,
			},
			want: nameWithId,
		},
		"TestWithoutSuffix": {
			args: args{
				suffix:  suffixID,
				replace: suffixId,
				name:    nameUnspecified,
			},
			want: nameUnspecified,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := suffixReplacer(tt.args.suffix, tt.args.replace)(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("suffixReplacer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReplacer(t *testing.T) {
	type args struct {
		old  string
		new  string
		name string
	}

	tests := map[string]struct {
		args args
		want string
	}{
		"TestMissing": {
			args: args{
				old:  prefixHTTP,
				new:  prefixHttp,
				name: nameUnspecified,
			},
			want: nameUnspecified,
		},
		"TestMatching": {
			args: args{
				old:  prefixHTTP,
				new:  prefixHttp,
				name: nameHTTPVersion,
			},
			want: nameHttpVersion,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Replacer(tt.args.old, tt.args.new)(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("replacer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapReplacer(t *testing.T) {
	type args struct {
		replaceMap map[string]string
		name       string
	}

	tests := map[string]struct {
		args args
		want string
	}{
		"TestMissing": {
			args: args{
				replaceMap: mapReplacerHTTPVersion,
				name:       nameUnspecified,
			},
			want: nameUnspecified,
		},
		"TestMatching": {
			args: args{
				replaceMap: mapReplacerHTTPVersion,
				name:       nameHTTPVersion,
			},
			want: nameHttpVersion,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := MapReplacer(tt.args.replaceMap)(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mapReplacer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLateInitializeFromResponse(t *testing.T) {
	type args struct {
		parentName     string
		crObject       interface{}
		responseObject interface{}
		opts           []LateInitOption
	}

	testStringCRField := "test-string-crField"
	testStringResponseField := "test-string-responseField"
	testInt64ResponseField := 1

	type nestedStruct1 struct {
		F1 *string
		F2 []*string
	}

	type nestedStruct2 struct {
		F1 *int
		F2 []*int
	}

	type nestedStruct3 struct {
		F1 *string
		F2 *string
	}

	type nestedStruct4 struct {
		F1 [][]*string
	}

	type nestedStruct5 struct {
		F1 [][]string
	}

	tests := map[string]struct {
		args         args
		wantModified bool
		wantErr      bool
		wantCRObject interface{}
	}{
		"NilCRObject": {
			args: args{
				responseObject: &struct{}{},
			},
		},
		"NilResponseObject": {
			args: args{
				crObject: &struct{}{},
			},
			wantCRObject: &struct{}{},
		},
		"TestNonStructCRObject": {
			args: args{
				crObject:       &testStringCRField,
				responseObject: &struct{}{},
			},
			wantErr: true,
		},
		"TestNonStructResponseObject": {
			args: args{
				crObject:       &struct{}{},
				responseObject: &testStringResponseField,
			},
			wantErr: true,
		},
		"TestEmptyStructCRAndResponseObjects": {
			args: args{
				crObject:       &struct{}{},
				responseObject: &struct{}{},
			},
			wantErr:      false,
			wantCRObject: &struct{}{},
		},
		"TestInitializedCRStringField": {
			args: args{
				crObject: &struct {
					F1 *string
				}{
					F1: &testStringCRField,
				},
				responseObject: &struct {
					F1 *string
				}{
					F1: &testStringResponseField,
				},
			},
			wantErr:      false,
			wantModified: false,
			wantCRObject: &struct {
				F1 *string
			}{
				F1: &testStringCRField,
			},
		},
		"TestUninitializedCRStringField": {
			args: args{
				crObject: &struct {
					F1 *string
				}{
					F1: nil,
				},
				responseObject: &struct {
					F1 *string
				}{
					F1: &testStringResponseField,
				},
			},
			wantErr:      false,
			wantModified: true,
			wantCRObject: &struct {
				F1 *string
			}{
				F1: &testStringResponseField,
			},
		},
		"TestInitializedCRNestedFields": {
			args: args{
				crObject: &struct {
					C1 *nestedStruct1
				}{
					C1: &nestedStruct1{
						F1: &testStringCRField,
						F2: []*string{
							&testStringCRField,
						},
					},
				},
				responseObject: &struct {
					C1 *nestedStruct1
				}{
					C1: &nestedStruct1{
						F1: &testStringResponseField,
						F2: []*string{
							&testStringResponseField,
						},
					},
				},
			},
			wantErr:      false,
			wantModified: false,
			wantCRObject: &struct {
				C1 *nestedStruct1
			}{
				C1: &nestedStruct1{
					F1: &testStringCRField,
					F2: []*string{
						&testStringCRField,
					},
				},
			},
		},
		"TestUninitializedCRNestedFields": {
			args: args{
				crObject: &struct {
					C1 *nestedStruct1
				}{
					C1: &nestedStruct1{},
				},
				responseObject: &struct {
					C1 *nestedStruct1
				}{
					C1: &nestedStruct1{
						F1: &testStringResponseField,
						F2: []*string{
							&testStringResponseField,
						},
					},
				},
			},
			wantErr:      false,
			wantModified: true,
			wantCRObject: &struct {
				C1 *nestedStruct1
			}{
				C1: &nestedStruct1{
					F1: &testStringResponseField,
					F2: []*string{
						&testStringResponseField,
					},
				},
			},
		},
		"TestFieldKindMismatch": {
			args: args{
				crObject: &nestedStruct1{
					F1: nil,
				},
				responseObject: &nestedStruct2{
					F1: &testInt64ResponseField,
				},
			},
			wantErr: true,
		},
		"TestNestedFieldKindMismatch": {
			args: args{
				crObject: &struct {
					C1 *nestedStruct1
				}{
					C1: &nestedStruct1{
						F1: nil,
					},
				},
				responseObject: &struct {
					C1 *nestedStruct2
				}{
					C1: &nestedStruct2{
						F1: &testInt64ResponseField,
					},
				},
			},
			wantErr: true,
		},
		"TestSliceItemKindMismatch": {
			args: args{
				crObject: &nestedStruct1{},
				responseObject: &nestedStruct3{
					F1: &testStringResponseField,
					F2: &testStringResponseField,
				},
			},
			wantErr: true,
		},
		"TestSliceOfSliceField": {
			args: args{
				crObject: &nestedStruct4{},
				responseObject: &nestedStruct4{
					F1: [][]*string{
						{
							&testStringResponseField,
						},
					},
				},
			},
			wantModified: true,
			wantCRObject: &nestedStruct4{
				F1: [][]*string{
					{
						&testStringResponseField,
					},
				},
			},
		},
		"TestUnsupportedSliceOfSliceField": {
			args: args{
				crObject: &nestedStruct5{},
				responseObject: &nestedStruct5{
					F1: [][]string{
						{
							testStringResponseField,
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := LateInitializeFromResponse(tt.args.parentName, tt.args.crObject, tt.args.responseObject, tt.args.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("lateInitializeFromResponse() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr {
				return
			}

			if got != tt.wantModified {
				t.Errorf("lateInitializeFromResponse() got = %v, want %v", got, tt.wantModified)
			}

			if diff := cmp.Diff(tt.wantCRObject, tt.args.crObject); diff != "" {
				t.Errorf("lateInitializeFromResponse(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	valTest := "testValue"
	valAnotherTest := "anotherTestValue"
	type args struct {
		actual  interface{}
		desired interface{}
		opts    []LateInitOption
	}
	tests := map[string]struct {
		args    args
		want    bool
		wantErr bool
	}{
		"SuccessNoOptsNoUpdate": {
			args: args{
				actual: &struct {
					S *string
				}{
					S: &valTest,
				},
				desired: &struct {
					S *string
				}{
					S: &valTest,
				},
			},
			want: true,
		},
		"SuccessNoOptsUpdateRequired": {
			args: args{
				actual: &struct {
					S *string
				}{
					S: &valAnotherTest,
				},
				desired: &struct {
					S *string
				}{
					S: &valTest,
				},
			},
		},
		"SuccessWithOptsNoUpdate": {
			args: args{
				actual: &struct {
					// linter disabled because we are testing a case based on
					// a common naming convention in aws-sdk-go
					Id *string //nolint:golint
				}{
					Id: &valTest,
				},
				desired: &struct {
					ID *string
				}{
					ID: &valTest,
				},
				opts: []LateInitOption{
					Replacer("ID", "Id"),
				},
			},
			want: true,
		},
		"FailNonPtr": {
			args: args{
				actual: &struct {
					S *string
				}{
					S: &valTest,
				},
				desired: struct {
					S *string
				}{
					S: &valTest,
				},
			},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, _, err := IsUpToDate(tt.args.actual, tt.args.desired, tt.args.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsUpToDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("IsUpToDate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
