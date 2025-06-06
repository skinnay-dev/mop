package hunter

import (
	"time"

	"github.com/wowsims/mop/sim/core"
)

func (hunter *Hunter) registerSerpentStingSpell() {

	hunter.SerpentSting = hunter.RegisterSpell(core.SpellConfig{
		ActionID:       core.ActionID{SpellID: 1978},
		SpellSchool:    core.SpellSchoolNature,
		ProcMask:       core.ProcMaskProc,
		ClassSpellMask: HunterSpellSerpentSting,
		Flags:          core.SpellFlagAPL,
		MissileSpeed:   40,
		MinRange:       0,
		MaxRange:       40,
		FocusCost: core.FocusCostOptions{
			Cost: 25,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: time.Second,
			},
			IgnoreHaste: true,
		},

		DamageMultiplierAdditive: 1,

		// SS uses Spell Crit which is multiplied by toxicology
		CritMultiplier:   hunter.CritMultiplier(1, 0),
		ThreatMultiplier: 1,

		Dot: core.DotConfig{
			Aura: core.Aura{
				ActionID: core.ActionID{SpellID: 1978},
				Label:    "SerpentStingDot",
				Tag:      "SerpentSting",
			},

			NumberOfTicks: 5,
			TickLength:    time.Second * 3,
			OnSnapshot: func(sim *core.Simulation, target *core.Unit, dot *core.Dot, isRollover bool) {
				baseDmg := 3240.4 + 0.032*dot.Spell.RangedAttackPower()
				dot.Snapshot(target, baseDmg)
			},
			OnTick: func(sim *core.Simulation, target *core.Unit, dot *core.Dot) {
				dot.CalcAndDealPeriodicSnapshotDamage(sim, target, dot.OutcomeTickPhysicalCrit)
			},
		},

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {

			result := spell.CalcOutcome(sim, target, spell.OutcomeRangedHit)

			spell.WaitTravelTime(sim, func(sim *core.Simulation) {
				if result.Landed() {
					spell.Dot(target).Apply(sim)
					// TODO: Check if hunter is survival specialization
					hunter.ImprovedSerpentSting.Cast(sim, target)
				}
				spell.DealOutcome(sim, result)
			})
		},
	})
}
