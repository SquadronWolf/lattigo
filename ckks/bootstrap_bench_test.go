package ckks

import (
	"math"
	"testing"
	"time"
)

func BenchmarkBootstrapp(b *testing.B) {

	if !*testBootstrapping {
		b.Skip("skipping bootstrapping test")
	}

	var err error
	var testContext = new(testParams)
	var btp *Bootstrapper

	paramSet := uint64(3)

	btpParams := DefaultBootstrapParams[paramSet]
	if testContext, err = genTestParams(DefaultBootstrapSchemeParams[paramSet], btpParams.H); err != nil {
		panic(err)
	}

	btpKey := testContext.kgen.GenBootstrappingKey(testContext.params.logSlots, btpParams, testContext.sk)

	if btp, err = NewBootstrapper(testContext.params, btpParams, btpKey); err != nil {
		panic(err)
	}

	b.Run(testString(testContext, "Bootstrapp/"), func(b *testing.B) {
		for i := 0; i < b.N; i++ {

			b.StopTimer()
			ct := NewCiphertextRandom(testContext.prng, testContext.params, 1, 0, testContext.params.scale)
			b.StartTimer()

			var t time.Time
			var ct0, ct1 *Ciphertext

			// Brings the ciphertext scale to Q0/2^{10}
			btp.evaluator.ScaleUp(ct, math.Round(btp.prescale/ct.Scale()), ct)

			// ModUp ct_{Q_0} -> ct_{Q_L}
			t = time.Now()
			ct = btp.modUp(ct)
			b.Log("After ModUp  :", time.Since(t), ct.Level(), ct.Scale())

			// Brings the ciphertext scale to sineQi/(Q0/scale) if its under
			btp.evaluator.ScaleUp(ct, math.Round(btp.postscale/ct.Scale()), ct)

			//SubSum X -> (N/dslots) * Y^dslots
			t = time.Now()
			ct = btp.subSum(ct)
			b.Log("After SubSum :", time.Since(t), ct.Level(), ct.Scale())

			// Part 1 : Coeffs to slots
			t = time.Now()
			ct0, ct1 = btp.coeffsToSlots(ct)
			b.Log("After CtS    :", time.Since(t), ct0.Level(), ct0.Scale())

			// Part 2 : SineEval
			t = time.Now()
			ct0, ct1 = btp.evaluateSine(ct0, ct1)
			b.Log("After Sine   :", time.Since(t), ct0.Level(), ct0.Scale())

			// Part 3 : Slots to coeffs
			t = time.Now()
			ct0 = btp.slotsToCoeffs(ct0, ct1)
			ct0.SetScale(math.Exp2(math.Round(math.Log2(ct0.Scale()))))
			b.Log("After StC    :", time.Since(t), ct0.Level(), ct0.Scale())
		}
	})
}
