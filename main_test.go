package main

import (
	"math"
	"testing"
)

func TestGetSimilarity(t *testing.T) {
	cases := []struct {
		name          string
		A             *Contact
		B             *Contact
		expected      float64
		expectedLabel string
	}{
		{
			name: "same email",
			A: &Contact{
				Email:   "foo@gmail.com",
				ZipCode: "baz",
				Address: "qux",
			},
			B: &Contact{
				FirstName: "baz",
				LastName:  "foo",
				Email:     "foo@gmail.com",
				ZipCode:   "qux",
				Address:   "bar",
			},
			expected:      1.0,
			expectedLabel: LABEL_ACCURACY_HIGH,
		},
		{
			name: "match full name, mismatch address",
			A: &Contact{
				FirstName: "baz",
				LastName:  "foo",
				ZipCode:   "baz",
				Address:   "qux",
			},
			B: &Contact{
				FirstName: "baz",
				LastName:  "foo",
				ZipCode:   "qux",
				Address:   "bar",
			},
			expectedLabel: LABEL_ACCURACY_LOW,
			expected:      weightFullName + (weightFullAddress * minSimAddress),
		},
		{
			name: "match full name, no more data",
			A: &Contact{
				FirstName: "baz",
				LastName:  "foo",
			},
			B: &Contact{
				FirstName: "baz",
				LastName:  "foo",
			},
			expectedLabel: LABEL_ACCURACY_MED,
			expected:      weightFullName,
		},
		{
			name: "full mismatch",
			A: &Contact{
				FirstName: "foo",
				LastName:  "bar",
				Email:     "foo1@gmail.com",
				ZipCode:   "baz",
				Address:   "qux",
			},
			B: &Contact{
				FirstName: "baz",
				LastName:  "foo",
				Email:     "foo2@gmail.com",
				ZipCode:   "qux",
				Address:   "bar",
			},
			expectedLabel: LABEL_ACCURACY_NHIGH,
			expected:      weightFullName*minSimName + weightFullAddress*minSimAddress + minSimEmail,
		},
		{
			name: "full match",
			A: &Contact{
				FirstName: "foo",
				LastName:  "bar",
				Email:     "foo1@gmail.com",
				ZipCode:   "baz",
				Address:   "qux",
			},
			B: &Contact{
				FirstName: "foo",
				LastName:  "bar",
				Email:     "foo1@gmail.com",
				ZipCode:   "baz",
				Address:   "qux",
			},
			expectedLabel: LABEL_ACCURACY_HIGH,
			expected:      1.0,
		},
		{
			name: "matches email mismatches rest",
			A: &Contact{
				FirstName: "foo1",
				LastName:  "bar1",
				Email:     "foo1@gmail.com",
				ZipCode:   "baz1",
				Address:   "qux1",
			},
			B: &Contact{
				FirstName: "foo",
				LastName:  "bar",
				Email:     "foo1@gmail.com",
				ZipCode:   "baz",
				Address:   "qux",
			},
			expectedLabel: LABEL_ACCURACY_HIGH,
			expected:      1.0,
		},
		{
			name: "matches full name and full address, email not given",
			A: &Contact{
				FirstName: "foo",
				LastName:  "bar",
				Email:     "foo1@gmail.com",
				ZipCode:   "baz",
				Address:   "qux",
			},
			B: &Contact{
				FirstName: "foo",
				LastName:  "bar",
				ZipCode:   "baz",
				Address:   "qux",
			},
			expectedLabel: LABEL_ACCURACY_HIGH,
			expected:      weightFullAddress + weightFullName,
		},
		{
			name: "matches only initials of the first and last names",
			A: &Contact{
				FirstName: "foo",
				LastName:  "b",
			},
			B: &Contact{
				FirstName: "f",
				LastName:  "bar",
			},
			expectedLabel: LABEL_ACCURACY_LOW,
			expected:      valPartialMatch * weightFullName * (weightFirstName + weightLastName),
		},
		{
			name: "matches only initials of the first name but mismatches last name",
			A: &Contact{
				FirstName: "foo",
				LastName:  "bar",
			},
			B: &Contact{
				FirstName: "f",
				LastName:  "baz",
			},
			expectedLabel: LABEL_ACCURACY_NMED,
			expected:      weightFullName * (-weightFirstName - weightLastName),
		},
		{
			name: "matches only initials of the last name but mismatches first name",
			A: &Contact{
				FirstName: "foo1",
				LastName:  "bar",
			},
			B: &Contact{
				FirstName: "foo2",
				LastName:  "b",
			},
			expectedLabel: LABEL_ACCURACY_NMED,
			expected:      weightFullName * (-weightFirstName - weightLastName),
		},
		{
			name:          "example inputs",
			A:             &Contact{FirstName: "C", LastName: "F", Email: "mollis.lectus.pede@outlook.net", Address: "449-6990 Tellus. Rd."},
			B:             &Contact{FirstName: "Ciara", LastName: "F", Email: "non.lacinia.at@zoho.ca", ZipCode: "39746"},
			expected:      valPartialMatch*weightFullName + minSimEmail,
			expectedLabel: LABEL_ACCURACY_LOW,
		},
		{
			name:          "empty contacts",
			A:             &Contact{},
			B:             &Contact{},
			expected:      0,
			expectedLabel: LABEL_ACCURACY_NULL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.A.getSimilarity(tc.B)
			if resultB := tc.B.getSimilarity(tc.A); !floatEq(result, resultB) {
				t.Errorf("getSimilarity(A, B) = %v mismatches getSimilarity(B, A) = %v", result, resultB)
			}
			if !floatEq(result, tc.expected) {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
			if labelScore(result) != tc.expectedLabel {
				t.Errorf("expected label %v, got %v for score %v", tc.expectedLabel, labelScore(result), result)
			}
		})
	}
}

func TestScore(t *testing.T) {
	tests := []struct {
		name     string
		contacts map[int]*Contact

		id         int
		wantScores map[int]float64
	}{
		{
			name: "example outputs #1",
			contacts: map[int]*Contact{
				1: {FirstName: "C", LastName: "F", Email: "mollis.lectus.pede@outlook.net", Address: "449-6990 Tellus. Rd."},
				2: {FirstName: "C", LastName: "French", Email: "mollis.lectus.pede@outlook.net", Address: "449-6990 Tellus. Rd.", ZipCode: "39746"},
				3: {FirstName: "Ciara", LastName: "F", Email: "non.lacinia.at@zoho.ca", ZipCode: "39746"},
			},

			id: 1,
			wantScores: map[int]float64{
				2: 1.0,
				3: 0.1,
			},
		},
		{
			name: "example outputs #2",
			contacts: map[int]*Contact{
				1: {FirstName: "C", LastName: "F", Email: "mollis.lectus.pede@outlook.net", Address: "449-6990 Tellus. Rd."},
				2: {FirstName: "C", LastName: "French", Email: "mollis.lectus.pede@outlook.net", Address: "449-6990 Tellus. Rd.", ZipCode: "39746"},
				3: {FirstName: "Ciara", LastName: "F", Email: "non.lacinia.at@zoho.ca", ZipCode: "39746"},
			},

			id: 2,
			wantScores: map[int]float64{
				3: 0.16,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Deduplicator{contacts: test.contacts}
			gotScores, err := d.score(test.id)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(gotScores) != len(test.wantScores) {
				t.Errorf("mismatched scores; want %v, got %v", test.wantScores, gotScores)
				return
			}

			for id, wantScore := range test.wantScores {
				gotScore, ok := gotScores[id]
				if !ok {
					t.Errorf("missing score for ID %v", id)
					continue
				}
				if !floatEq(gotScore, wantScore) {
					t.Errorf("ID %v: got %v, want %v", id, gotScore, wantScore)
				}
			}
		})
	}
}

// as tests expected values are defined with constants, Go will define precise values on build time, while the others
// are defined in runtime, leading to potential differences (e.g: -0.1375 vs -0.13749999999998)
func floatEq(a, b float64) bool {
	epsilon := 1e-9
	return math.Abs(a-b) <= epsilon
}
