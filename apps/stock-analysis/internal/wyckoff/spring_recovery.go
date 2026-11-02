package wyckoff

import "time"

// SpringAttempt records both failures and delayed recovery against bounds that
// were already available. A recovered Spring remains a market hypothesis.
type SpringAttempt struct {
	Date                      time.Time  `json:"date"`
	LowestLow                 float64    `json:"lowest_low"`
	ObservedAt                time.Time  `json:"observed_at"`
	RecoveryDate              *time.Time `json:"recovery_date,omitempty"`
	Status                    string     `json:"status"`
	Reason                    string     `json:"reason"`
	startIndex, recoveryIndex int
}

func (s *StructureState) advanceSpring(cs []Candle, i int) {
	c := cs[i]
	width := s.Resistance - s.Support
	newAttempt:=len(s.SpringAttempts)==0
 if !newAttempt && i>=2 {
  previous:=s.SpringAttempts[len(s.SpringAttempts)-1]
  terminal:=previous.Status=="EXPIRED"||previous.Status=="REJECTED"||previous.Status=="INVALIDATED"
  independent:=previous.Status=="RECOVERED"&&i>previous.recoveryIndex+15
  newAttempt=(terminal||independent)&&c.Low<s.Support&&cs[i-1].Close>=s.Support&&cs[i-2].Close>=s.Support
 }
 if newAttempt {
		if s.Status != "RANGE" || s.BreakoutDate != nil || c.Low >= s.Support {
			return
		}
		a := SpringAttempt{Date: c.Date, LowestLow: c.Low, ObservedAt: c.Date, Status: "PENDING", Reason: "Support penetrated; wait for recovery within three subsequent observed sessions.", startIndex: i, recoveryIndex: -1}
		if s.Support-c.Low > width {
			a.Status = "REJECTED"
			a.Reason = "Penetration exceeded the original range width; do not infer Spring from a deep breakdown."
		}
		s.SpringAttempts = append(s.SpringAttempts, a)
	}
	a := &s.SpringAttempts[len(s.SpringAttempts)-1]
	if a.Status == "RECOVERED" {
		if c.Low < a.LowestLow {
			a.Status = "INVALIDATED"
			a.ObservedAt = c.Date
			a.Reason = "A lower low after recovery invalidated this Spring/Test sequence."
			return
		}
		if i > a.recoveryIndex+15 {
			return
		}
		if i <= a.recoveryIndex || i < a.startIndex+2 {
			return
		}
		for _, e := range s.Events {
			if e.Type == "TEST" && e.Date.After(a.Date) {
				return
			}
		}
		st := stats(c, avgVolume(cs[maxInt(0, i-20):i]))
		if c.Close >= s.Support && c.Low <= s.Support+width*.2 && st.relativeVolume < .8 && st.closeLocation >= .55 {
			s.Events=append(s.Events,Event{Type:"TEST",Date:c.Date,Confidence:.6,StructureID:s.ID,Evidence:"Candidate Test after recovered Spring held its low and original support, with reduced supply."})
			for j := range s.Events {
				if s.Events[j].Type == "TEST" && s.Events[j].Date.Equal(c.Date) {
					d := c.Date
					s.Events[j].ObservedAt = &d
					s.Events[j].Status = "candidate"
				}
			}
		}
		return
	}
	if a.Status != "PENDING" {
		return
	}
	a.ObservedAt = c.Date
	if c.Low < a.LowestLow {
		a.LowestLow = c.Low
	}
	if s.Support-a.LowestLow > width {
		a.Status = "REJECTED"
		a.Reason = "Penetration exceeded the original range width before recovery."
		return
	}
	if i-a.startIndex > 3 {
		a.Status = "EXPIRED"
		a.Reason = "No qualifying recovery within three subsequent observed sessions."
		return
	}
	st := stats(c, avgVolume(cs[maxInt(0, i-20):i]))
	// Same-session low-volume recovery and multi-session demand recovery are
	// separate paths. High-volume penetration is not automatically rejected.
	recovered := c.Close > s.Support && st.closeLocation >= .5
	if i == a.startIndex {
		recovered = recovered && st.relativeVolume < 1.35
	}
	if i > a.startIndex {
		recovered = recovered && c.Close > cs[i-1].Close
	}
	if !recovered {
		return
	}
	d := c.Date
	a.RecoveryDate = &d
	a.recoveryIndex = i
	a.Status = "RECOVERED"
	a.Reason = "Price recovered above pre-existing support with an improved close; candidate Spring, not proof of accumulation."
	s.Events=append(s.Events,Event{Type:"SPRING",Date:a.Date,Confidence:.6,StructureID:s.ID,Evidence:a.Reason})
	for j := range s.Events {
		if s.Events[j].Type == "SPRING" && s.Events[j].Date.Equal(a.Date) {
			s.Events[j].ObservedAt = &d
			s.Events[j].Status = "candidate"
		}
	}
}
