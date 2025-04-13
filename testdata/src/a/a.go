package a

import "fmt"

// TODO: Comments should be ignored?
// TODO: When counting lines, ignore `return`?

// NOTE: All of the following count as a return statement:
// - `return ...`
// - `panic(...)`
// - `os.Exit(...)`
// - End of function

// In comments below:
// - body_1 is "then body of conditionOne"
// - after_1 is "the code that goes after the conditionOne"

// Test1
// ❌conditionOne should be inverted:
//   - YES: after_1 (1 line) is shorter than body_1 (3 lines)
//   - YES: both body_1 and after_1 end with a return statement
func Test1(conditionOne bool) {
	if conditionOne { // want "consider inverting the condition to return early"
		fmt.Println("a")
		fmt.Println("b")
		fmt.Println("c")
		return
	}
	fmt.Println("d")
}

// Test2
// ✅conditionOne should NOT be inverted:
//   - NO: body_1 has no return statement (both body_1 should always be executed, and always in this specific order).
func Test2(conditionOne bool) {
	if conditionOne {
		fmt.Println("a")
		fmt.Println("b")
		fmt.Println("c")
	}
	fmt.Println("d")
}

// Test3
// ❌conditionOne should be inverted:
//   - YES: body_1 has no return statement, but after_1 is empty
//   - YES: after_1 (0 lines) is shorter than body_1 (3 lines)
func Test3(conditionOne bool) {
	if conditionOne { // want "consider inverting the condition to return early"
		fmt.Println("a")
		fmt.Println("b")
		fmt.Println("c")
	}
}

// Test4
// ❌conditionOne should be inverted:
//   - YES: after_1 (1 line) is shorter than body_1 (6 lines)
//   - YES: all branches in body_1 end up with return statement, after_1 ends with a return statement
// ❌conditionTwo should be inverted:
//   - YES: after_2 (1 line) is shorter than body_2 (3 lines)
//   - YES: both body_2 and after_2 end with a return statement
func Test4(conditionOne bool, conditionTwo bool) {
	if conditionOne { // want "consider inverting the condition to return early"
		if conditionTwo { // want "consider inverting the condition to return early"
			fmt.Println("a")
			fmt.Println("b")
			fmt.Println("c")
			return
		}
		fmt.Println("d")
		return
	}
	fmt.Println("e")
}

// Test5
// ✅conditionOne should NOT be inverted:
//   - NO: not all branches in body_1 end with a return statement
//   - NO: after_1 (6 lines) is longer than body_1 (4 lines)
// ✅conditionTwo should NOT be inverted:
//   - NO: none of body_2 and after_2 end with a return statement (both should)
func Test5(conditionOne bool, conditionTwo bool) {
	if conditionOne {
		if conditionTwo {
			fmt.Println("a")
			fmt.Println("b")
		}
		fmt.Println("c")
	}
	fmt.Println("d")
	fmt.Println("e")
	fmt.Println("f")
	fmt.Println("g")
	fmt.Println("h")
	fmt.Println("i")
}

// Test6
// ✅ conditionOne should NOT be inverted:
//   - NO: after_1 (6 lines) is longer than after_1 (5 line)
// ❌ conditionTwo should be inverted:
//   - YES: after_2 (1 line) is shorter than body_2 (2 lines)
//   - YES: both body_2 and after_2 end with a return statement
func Test6(conditionOne bool, conditionTwo bool) {
	if conditionOne {
		if conditionTwo { // want "consider inverting the condition to return early"
			fmt.Println("a")
			fmt.Println("b")
			return
		}
		fmt.Println("c")
		return
	}
	fmt.Println("d")
	fmt.Println("e")
	fmt.Println("f")
	fmt.Println("g")
	fmt.Println("h")
	fmt.Println("i")
}

// Test7
// ❌conditionOne should be inverted:
//   - YES: after_1 (1 line) is shorter than body_1 (5 lines)
//   - YES: all branches in body_1 end up with return statement, after_1 ends with a return statement
// ✅condition_2 should NOT be inverted
// - body_2 has no return and should always be executed before after_2
func Test7(conditionOne bool, conditionTwo bool) {
	if conditionOne { // want "consider inverting the condition to return early"
		if conditionTwo {
			fmt.Println("a")
		}
		fmt.Println("b")
		fmt.Println("c")
		return
	}
	fmt.Println("d")
}

// Test8
// ❌conditionOne should be inverted:
//   - YES: after_1 (1 line) is shorter than body_1 (8 lines)
//   - YES: all branches in body_1 end up with a return statement, after_1 ends with a return statement
// ✅conditionTwo should NOT be inverted:
//   - NO: body_2 has no return and should always be executed before after_2
// ✅conditionThree should NOT be inverted:
//   - YES: after_3 (1 line) is shorter than body_3 (3 lines)
//   - NO: after_3 has no return statement
func Test8(conditionOne bool, conditionTwo bool, conditionThree bool) {
	if conditionOne { // want "consider inverting the condition to return early"
		if conditionTwo {
			if conditionThree {
				fmt.Println("a")
				fmt.Println("b")
				fmt.Println("c")
				return
			}
			fmt.Println("d")
		}
		fmt.Println("e")
		fmt.Println("f")
		fmt.Println("g")
		return
	}
	fmt.Println("h")
}

// Test9
// ❌conditionOne should be inverted:
//   - YES: after_1 (1 line) is shorter than body_1 (8 lines)
//   - YES: all branches in body_1 end up with a return statement, after_1 ends with a return statement
// ❌conditionTwo should NOT be inverted:
//   - YES: after_2 (4 lines) is shorter than body_2 (8 lines)
//   - YES: both body_2 and after_2 have return statements
// ❌conditionThree should be inverted:
//   - YES: after_3 (1 line) is shorter than body_3 (3 lines)
//   - YES: both body_3 and after_3 end with a return statement
func Test9(conditionOne bool, conditionTwo bool, conditionThree bool) {
	if conditionOne { // want "consider inverting the condition to return early"
		if conditionTwo { // want "consider inverting the condition to return early"
			if conditionThree { // want "consider inverting the condition to return early"
				fmt.Println("a")
				fmt.Println("b")
				fmt.Println("c")
				return
			}
			fmt.Println("d")
			return
		}
		fmt.Println("e")
		fmt.Println("f")
		fmt.Println("g")
		return
	}
	fmt.Println("h")
}

// Test10
// ✅conditionOne should NOT be inverted:
//   - NO: body_1 has no return statement (both body_1 should always be executed, and always in this specific order).
func Test10(conditionOne bool) bool {
	if conditionOne {
		fmt.Println("a")
		fmt.Println("b")
	}
	return true
}

// Test11
func Test11(conditionOne bool) {
	if conditionOne {
		return
	}
	return // panic() // os.Exit(1)
}

// TODO: Support `continue` and `break` statements
