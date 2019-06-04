package exporter

import (
	"time"
)

const (
	adcVref = 33.0
)

type batteryState struct {
	currentAmps         float64
	energyMilliampHours float64
	voltageVolt         float64
	lastTime            int64
	currentScale        int32
	currentOffset       int64
	vbatScale           int64
}

func (b *batteryState) setLatestAmperage(value int64, newTime int64) {
	b.currentAmps = float64(value*adcVref*100/4095-b.currentOffset) * 10 / float64(b.currentScale)
	if b.lastTime != 0.0 {
		b.energyMilliampHours += b.currentAmps * float64(newTime-b.lastTime) / time.Hour.Seconds() / 1000
	}
	b.lastTime = newTime
}

func (b *batteryState) setLatestVbat(value int64) {
	// ADC is 12 bit (i.e. max 0xFFF), voltage reference is 3.3V, vbatscale is premultiplied by 100
	b.voltageVolt = float64(value*adcVref*b.vbatScale) / 4095 / 100.0
}
