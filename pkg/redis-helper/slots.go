package rh

import (
	"fmt"
	"strconv"
	"strings"
)

type Slots []bool

type SlotSlice struct {
	Begin int
	End   int
}

func (slice *SlotSlice) parse(s string) error {
	if s == "" {
		return nil
	}

	nums := strings.SplitN(s, "-", 2)

	slot, err := strconv.Atoi(nums[0])
	if err != nil {
		return err
	}
	slice.Begin = slot

	// nolint
	if len(nums) < 2 {
		slice.End = slice.Begin
		return nil
	}

	slot, err = strconv.Atoi(nums[1])
	if err != nil {
		return err
	}
	slice.End = slot

	return nil
}

func NewSlots() Slots {
	return make([]bool, TotalSlots)
}

func (p Slots) Set(slot int) error {
	if slot < 0 || slot >= TotalSlots {
		return fmt.Errorf("slot %d out of range", slot)
	}
	if p[slot] {
		return fmt.Errorf("slot %d already set", slot)
	}
	p[slot] = true
	return nil
}

func (p Slots) Unset(slot int) error {
	if slot < 0 || slot >= TotalSlots {
		return fmt.Errorf("slot %d out of range", slot)
	}
	if !p[slot] {
		return fmt.Errorf("slot %d already unset", slot)
	}

	p[slot] = false
	return nil
}

func (p Slots) IsSet(slot int) bool {
	return p[slot]
}

func (p Slots) IsUnset(slot int) bool {
	return !p[slot]
}

func (p Slots) IsEmpty() bool {
	for _, s := range p {
		if s {
			return false
		}
	}

	return true
}

func (p Slots) IsAllSet() bool {
	for _, s := range p {
		if !s {
			return false
		}
	}

	return true
}

func (p Slots) SetSlotSlice(slotSlice string) error {
	slice := SlotSlice{}
	err := slice.parse(slotSlice)
	if err != nil {
		return err
	}

	for i := slice.Begin; i <= slice.End; i++ {
		err = p.Set(i)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p Slots) SlotsCount() int {
	count := 0
	for _, s := range p {
		if s {
			count++
		}
	}

	return count
}

func (p Slots) String() string {
	var result strings.Builder
	length := TotalSlots
	inRange := false
	start := 0

	for i := 0; i < length; i++ {
		if p[i] {
			if !inRange {
				inRange = true
				start = i
			}
			// Handle the case where the last element is '*'
			if i == length-1 {
				appendSlotRange(&result, start, i)
			}
		} else {
			if inRange {
				appendSlotRange(&result, start, i-1)
				inRange = false
			}
		}
	}

	return strings.TrimSpace(result.String())
}

// appendSlotRange handles and appends the range string to the result
func appendSlotRange(result *strings.Builder, start, end int) {
	if start == end {
		result.WriteString(strconv.Itoa(start) + " ")
	} else {
		result.WriteString(strconv.Itoa(start) + "-" + strconv.Itoa(end) + " ")
	}
}

func (p Slots) Compare(other Slots) (bool, string) {
	leakSlots := NewSlots()
	extraSlots := NewSlots()

	for i := 0; i < TotalSlots; i++ {
		if p[i] && !other[i] {
			extraSlots[i] = true
		} else if !p[i] && other[i] {
			leakSlots[i] = true
		}
	}

	if leakSlots.IsEmpty() && extraSlots.IsEmpty() {
		return true, ""
	}

	return false, fmt.Sprintf("-[%s] +[%s]", leakSlots.String(), extraSlots.String())
}
