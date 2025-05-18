package paladin

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

func (paladin *Paladin) registerDivineProtectionSpell() {
	glyphOfDivineProtection := paladin.HasMajorGlyph(proto.PaladinMajorGlyph_GlyphOfDivineProtection)

	spellDamageMultiplier := core.TernaryFloat64(glyphOfDivineProtection, 0.8, 0.6)
	physDamageMultiplier := core.TernaryFloat64(glyphOfDivineProtection, 0.8, 1.0)

	actionID := core.ActionID{SpellID: 498}
	paladin.DivineProtectionAura = paladin.RegisterAura(core.Aura{
		Label:    "Divine Protection" + paladin.Label,
		ActionID: actionID,
		Duration: time.Second * 10,

		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexArcane] *= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] *= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] *= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly] *= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] *= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] *= spellDamageMultiplier
			paladin.PseudoStats.DamageTakenMultiplier *= physDamageMultiplier
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexArcane] /= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFire] /= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexFrost] /= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexHoly] /= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexNature] /= spellDamageMultiplier
			paladin.PseudoStats.SchoolDamageTakenMultiplier[stats.SchoolIndexShadow] /= spellDamageMultiplier
			paladin.PseudoStats.DamageTakenMultiplier /= physDamageMultiplier
		},
	})

	paladin.DivineProtection = paladin.RegisterSpell(core.SpellConfig{
		ActionID:       actionID,
		Flags:          core.SpellFlagAPL,
		ClassSpellMask: SpellMaskDivineProtection,

		ManaCost: core.ManaCostOptions{
			BaseCostPercent: 3.5,
		},
		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				NonEmpty: true,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: time.Minute * 1,
			},
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, spell *core.Spell) {
			paladin.DivineProtectionAura.Activate(sim)
		},
	})

	if paladin.Spec == proto.Spec_SpecProtectionPaladin {
		paladin.AddMajorCooldown(core.MajorCooldown{
			Spell: paladin.DivineProtection,
			Type:  core.CooldownTypeSurvival,
		})
	}
}
