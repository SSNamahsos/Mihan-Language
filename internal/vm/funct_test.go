package vm

import (
	"bytes"
	"strings"
	"testing"
)

func TestAFunctionAgreesWithTheLanguage(t *testing.T) {
	runOnBoth(t, `
funct jam(a, b) -> arz
    bargard a + b
chap_khat jam(2, 3)
chap_khat jam(10, 20)
`)
}

func TestAFunctionWithNoArgumentsAgrees(t *testing.T) {
	runOnBoth(t, `
funct salam()
    bargard "salam"
funct donbale(k)
    chap_khat k
    chap_khat k
    chap_khat k
donbale(salam())
`)
}

func TestRecursionAgrees(t *testing.T) {
	runOnBoth(t, `
funct fib(n) -> arz
    agar n < 2
        bargard n
    bargard fib(n - 1) + fib(n - 2)
har i az 0 ta 20
    chap_khat fib(i)
`)
}

func TestDeepRecursionAgrees(t *testing.T) {
	runOnBoth(t, `
funct down(n) -> arz
    agar n == 0
        bargard 0
    bargard 1 + down(n - 1)
chap_khat down(500)
`)
}

func TestAFunctionsNamesAreItsOwn(t *testing.T) {
	runOnBoth(t, `
funct one()
    x = 10
    bargard x
funct two()
    x = 20
    bargard x
chap_khat one()
chap_khat two()
`)
}

func TestAClosureMovesWhatItsCallerCanSee(t *testing.T) {
	runOnBoth(t, `
funct shomar()
    adad = 0
    funct she()
        adad = adad + 1
        bargard adad
    she()
    she()
    she()
    bargard adad
chap_khat shomar()
`)
}

func TestTwoClosuresShareOneName(t *testing.T) {
	runOnBoth(t, `
funct amaliat()
    adad = 0
    funct ziad()
        adad = adad + 1
        bargard adad
    funct begam()
        adad = adad + 10
        bargard adad
    ziad()
    ziad()
    ziad()
    begam()
    ziad()
    bargard adad
chap_khat amaliat()
`)
}

func TestAClosureOutlivesTheCallThatMadeIt(t *testing.T) {
	runOnBoth(t, `
funct sazar() -> funct
    adad = 100
    funct she()
        adad = adad + 1
        bargard adad
    bargard she
shomar = sazar()
chap_khat shomar()
chap_khat shomar()
chap_khat shomar()
`)
}

func TestAFunctionInACollectionIsRefusedNotGuessed(t *testing.T) {
	src := `
funct jam(a, b) -> arz
    bargard a * b
har shoma dar [jam]
    chap_khat shoma(6, 3)
`
	p := lowerFor(t, src)
	var out bytes.Buffer
	err := New(p, &out).Run()
	if err == nil {
		t.Fatal("the machine ran a function it had lost the instructions for, which it must not do")
	}
	if !strings.Contains(err.Error(), "call") {
		t.Errorf("the machine should have said the value cannot be called, and said: %v", err)
	}
}

func TestAFunctionAsAnArgumentAgrees(t *testing.T) {
	runOnBoth(t, `
funct zard(a, b) -> arz
    bargard a - b
funct begoz(f, a, b) -> arz
    bargard f(a, b)
chap_khat begoz(zard, 9, 4)
chap_khat begoz(zard, 1, 2)
`)
}

func TestManyCallsDoNotGrowTheMachine(t *testing.T) {
	p := lowerFor(t, `
funct jam(a, b) -> arz
    bargard a + b
n = 0
har i az 0 ta 200_000
    n = jam(n, 1)
chap_khat n
`)
	m := New(p, &devNull{})
	before := m.usedSlots()
	if err := m.Run(); err != nil {
		t.Fatal(err)
	}
	if after := m.usedSlots(); after > before+1 {
		t.Errorf("the machine kept %d slots after 200,000 calls, so a call is not giving its locals back", after-before)
	}
}
