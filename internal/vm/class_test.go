package vm

import (
	"testing"
)

func TestAClassAgrees(t *testing.T) {
	runOnBoth(t, `
darghan Ostad
    baza naam
    baza sen
    kiloh = 0

    __sakht__(naam, sen)
        khod.naam = naam
        khod.sen = sen

    barkhordan()
        chap_khat khod.naam

    jashoor()
        chap_khat $"I am {khod.naam} and I am {khod.sen}"

o = Ostad("Sara", 25)
chap_khat o.naam
chap_khat o.sen
o.kiloh = 5
chap_khat o.kiloh
o.barkhordan()
o.jashoor()
`)
}

func TestInheritanceAgrees(t *testing.T) {
	runOnBoth(t, `
darghan Ostad
    baza naam
    kiloh = 0

    __sakht__(naam)
        khod.naam = naam

    salom()
        chap_khat $"man {khod.naam} hastam"

darghan Ostad_Baftshenas < Ostad>
    reshte()
        chap_khat $"Ostad says {khod.kiloh}"

b = Ostad_Baftshenas("Ali")
chap_khat b.naam
b.kiloh = 5
b.salom()
b.reshte()
`)
}

func TestAnEnumAgrees(t *testing.T) {
	runOnBoth(t, `
shenase Rang
    sefid = 0
    siah = 1
    sabz = 7

chap_khat Rang.sefid
chap_khat Rang.siah
chap_khat Rang.sabz
`)
}

func TestAStructAgrees(t *testing.T) {
	runOnBoth(t, `
sakht Noghat
    x = 0
    y = 0

n = Noghat()
n.x = 1
n.y = 2
chap_khat n.x
chap_khat n.y
`)
}

func TestFieldTypeZerosAgree(t *testing.T) {
	runOnBoth(t, `
darghan Jabe
    shomaresh: arz
    naam: matn
    arzesh: kasr
    cheshm = "sabz"

j = Jabe()
chap_khat j.shomaresh
chap_khat j.naam
chap_khat j.arzesh
chap_khat j.cheshm
`)
}

func TestAMethodCallingAMethodAgrees(t *testing.T) {
	runOnBoth(t, `
darghan Hesab
    adad = 0

    do_barabar()
        bargard khod.adad * 2

    neshan_kon()
        chap_khat khod.do_barabar()

h = Hesab()
h.adad = 21
h.neshan_kon()
`)
}

func TestMethodsWithArgumentsAgree(t *testing.T) {
	runOnBoth(t, `
darghan Jam
    kol = 0

    ziad(n)
        khod.kol += n
        bargard khod.kol

j = Jam()
chap_khat j.ziad(5)
chap_khat j.ziad(7)
chap_khat j.kol
`)
}

func TestAFieldWrittenInALoopDoesNotGrowTheMachine(t *testing.T) {
	p := lowerFor(t, `
darghan P
    x = 0
p = P()
i = 0
ta i < 100_000
    p.x = i
    i += 1
`)
	m := New(p, &devNull{})
	if err := m.Run(); err != nil {
		t.Fatal(err)
	}
	if got := len(m.stack); got > 10 {
		t.Errorf("the stack holds %d values after 100,000 writes, so a write is leaking", got)
	}
	if got := m.usedSlots(); got > 10 {
		t.Errorf("the machine kept %d slots after 100,000 writes", got)
	}
}

func TestAListSlotWrittenInALoopDoesNotGrowTheMachine(t *testing.T) {
	p := lowerFor(t, `
xs = [0]
i = 0
ta i < 100_000
    xs[0] = i
    i += 1
`)
	m := New(p, &devNull{})
	if err := m.Run(); err != nil {
		t.Fatal(err)
	}
	if got := len(m.stack); got > 10 {
		t.Errorf("the stack holds %d values after 100,000 writes, so a write is leaking", got)
	}
}

func TestAClassInsideAFunctionAgrees(t *testing.T) {
	runOnBoth(t, `
funct sazar(n)
    darghan Kuti
        andaze = 0
    k = Kuti()
    k.andaze = n
    bargard k.andaze

chap_khat sazar(3)
chap_khat sazar(4)
`)
}
