package hunter

import (
	"time"

	"github.com/wowsims/mop/sim/core"
	"github.com/wowsims/mop/sim/core/proto"
	"github.com/wowsims/mop/sim/core/stats"
)

const ThoridalTheStarsFuryItemID = 34334

type Hunter struct {
	core.Character

	ClassSpellScaling float64

	Talents             *proto.HunterTalents
	Options             *proto.HunterOptions
	BeastMasteryOptions *proto.BeastMasteryHunter_Options
	MarksmanshipOptions *proto.MarksmanshipHunter_Options
	SurvivalOptions     *proto.SurvivalHunter_Options

	// Pet *HunterPet

	// The most recent time at which moving could have started, for trap weaving.
	mayMoveAt time.Duration

	AspectOfTheHawk *core.Spell
	AspectOfTheFox  *core.Spell

	FireTrapTimer *core.Timer

	// Hunter spells
	KillCommand   *core.Spell
	ArcaneShot    *core.Spell
	ExplosiveTrap *core.Spell
	KillShot      *core.Spell
	RapidFire     *core.Spell
	MultiShot     *core.Spell
	RaptorStrike  *core.Spell
	SerpentSting  *core.Spell
	SteadyShot    *core.Spell
	ScorpidSting  *core.Spell
	SilencingShot *core.Spell
	TrapLauncher  *core.Spell

	// BM only spells

	// MM only spells
	AimedShot   *core.Spell
	ChimeraShot *core.Spell

	// Survival only spells
	ExplosiveShot *core.Spell
	BlackArrow    *core.Spell
	CobraShot     *core.Spell

	// Fake spells to encapsulate weaving logic.
	TrapWeaveSpell                *core.Spell
	ImprovedSerpentSting          *core.Spell
	AspectOfTheHawkAura           *core.StatBuffAura
	AspectOfTheFoxAura            *core.Aura
	ImprovedSteadyShotAura        *core.Aura
	ImprovedSteadyShotAuraCounter *core.Aura
	LockAndLoadAura               *core.Aura
	RapidFireAura                 *core.Aura
	ScorpidStingAuras             core.AuraArray
	KillingStreakCounterAura      *core.Aura
	KillingStreakAura             *core.Aura
	MasterMarksmanAura            *core.Aura
	MasterMarksmanCounterAura     *core.Aura
	TrapLauncherAura              *core.Aura

	// Item sets
	T13_2pc *core.Aura
}

func (hunter *Hunter) GetCharacter() *core.Character {
	return &hunter.Character
}

func (hunter *Hunter) GetHunter() *Hunter {
	return hunter
}

func NewHunter(character *core.Character, options *proto.Player, hunterOptions *proto.HunterOptions) *Hunter {
	hunter := &Hunter{
		Character:         *character,
		Talents:           &proto.HunterTalents{},
		Options:           hunterOptions,
		ClassSpellScaling: core.GetClassSpellScalingCoefficient(proto.Class_ClassHunter),
	}

	core.FillTalentsProto(hunter.Talents.ProtoReflect(), options.TalentsString)
	focusPerSecond := 4.0

	// TODO: Fix this to work with the new talent system.
	// hunter.EnableFocusBar(100+(float64(hunter.Talents.KindredSpirits)*5), focusPerSecond, true, nil)
	hunter.EnableFocusBar(100, focusPerSecond, true, nil)

	hunter.PseudoStats.CanParry = true

	// Passive bonus (used to be from quiver).
	//hunter.PseudoStats.RangedSpeedMultiplier *= 1.15
	rangedWeapon := hunter.WeaponFromRanged(0)

	hunter.EnableAutoAttacks(hunter, core.AutoAttackOptions{
		Ranged: rangedWeapon,
		//ReplaceMHSwing:  hunter.TryRaptorStrike, //Todo: Might be weaving
		AutoSwingRanged: true,
		AutoSwingMelee:  false,
	})

	hunter.AutoAttacks.RangedConfig().ApplyEffects = func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
		baseDamage := hunter.RangedWeaponDamage(sim, spell.RangedAttackPower())

		result := spell.CalcDamage(sim, target, baseDamage, spell.OutcomeRangedHitAndCrit)

		spell.WaitTravelTime(sim, func(sim *core.Simulation) {
			spell.DealDamage(sim, result)
		})
	}

	hunter.AddStatDependencies()
	// hunter.Pet = hunter.NewHunterPet()
	return hunter
}

func (hunter *Hunter) Initialize() {
	hunter.AutoAttacks.MHConfig().CritMultiplier = hunter.DefaultCritMultiplier()
	hunter.AutoAttacks.OHConfig().CritMultiplier = hunter.DefaultCritMultiplier()
	hunter.AutoAttacks.RangedConfig().CritMultiplier = hunter.DefaultCritMultiplier()

	hunter.FireTrapTimer = hunter.NewTimer()

	// hunter.ApplyGlyphs()
	hunter.RegisterSpells()

	// hunter.addBloodthirstyGloves()
}

func (hunter *Hunter) ApplyTalents() {}

func (hunter *Hunter) RegisterSpells() {
	// hunter.registerSteadyShotSpell()
	hunter.registerArcaneShotSpell()
	hunter.registerKillShotSpell()
	// hunter.registerAspectOfTheHawkSpell()
	hunter.registerSerpentStingSpell()
	// hunter.registerMultiShotSpell()
	// hunter.registerKillCommandSpell()
	// hunter.registerExplosiveTrapSpell(hunter.FireTrapTimer)
	hunter.registerCobraShotSpell()
	// hunter.registerRapidFireCD()
	// hunter.registerSilencingShotSpell()
	hunter.registerRaptorStrikeSpell()
	// hunter.registerTrapLauncher()
	hunter.registerHuntersMarkSpell()
	// hunter.registerAspectOfTheFoxSpell()
}

func (hunter *Hunter) AddStatDependencies() {
	hunter.AddStatDependency(stats.Agility, stats.AttackPower, 2)
	hunter.AddStatDependency(stats.Agility, stats.RangedAttackPower, 2)
	hunter.AddStatDependency(stats.Agility, stats.PhysicalCritPercent, core.CritPerAgiMaxLevel[hunter.Class])
}

func (hunter *Hunter) AddRaidBuffs(raidBuffs *proto.RaidBuffs) {
	// TODO: Fix this to work with the new talent system.
	// if hunter.Talents.TrueshotAura {
	// 	raidBuffs.TrueshotAura = true
	// }
	// if hunter.Talents.FerociousInspiration && hunter.Options.PetType != proto.HunterOptions_PetNone {
	// 	raidBuffs.FerociousInspiration = true
	// }

	if hunter.Options.PetType == proto.HunterOptions_CoreHound {
		raidBuffs.Bloodlust = true
	}

	if hunter.Options.PetType == proto.HunterOptions_ShaleSpider {
		raidBuffs.BlessingOfKings = true
	}

	if hunter.Options.PetType == proto.HunterOptions_Wolf || hunter.Options.PetType == proto.HunterOptions_Devilsaur {
		raidBuffs.FuriousHowl = true
	}

	// TODO: Fix this to work with the new talent system.
	//
	//	if hunter.Talents.HuntingParty {
	//		raidBuffs.HuntingParty = true
	//	}
}

func (hunter *Hunter) AddPartyBuffs(_ *proto.PartyBuffs) {
}

func (hunter *Hunter) HasMajorGlyph(glyph proto.HunterMajorGlyph) bool {
	return hunter.HasGlyph(int32(glyph))
}
func (hunter *Hunter) HasMinorGlyph(glyph proto.HunterMinorGlyph) bool {
	return hunter.HasGlyph(int32(glyph))
}

func (hunter *Hunter) Reset(_ *core.Simulation) {
	hunter.mayMoveAt = 0
}

const (
	HunterSpellFlagsNone int64 = 0
	SpellMaskSpellRanged int64 = 1 << iota
	HunterSpellAutoShot
	HunterSpellSteadyShot
	HunterSpellCobraShot
	HunterSpellArcaneShot
	HunterSpellKillCommand
	HunterSpellChimeraShot
	HunterSpellExplosiveShot
	HunterSpellExplosiveTrap
	HunterSpellBlackArrow
	HunterSpellMultiShot
	HunterSpellAimedShot
	HunterSpellSerpentSting
	HunterSpellKillShot
	HunterSpellRapidFire
	HunterSpellBestialWrath
	HunterPetFocusDump
	HunterSpellsTierTwelve = HunterSpellArcaneShot | HunterSpellKillCommand | HunterSpellChimeraShot | HunterSpellExplosiveShot |
		HunterSpellMultiShot | HunterSpellAimedShot
	HunterSpellsAll = HunterSpellSteadyShot | HunterSpellCobraShot |
		HunterSpellArcaneShot | HunterSpellKillCommand | HunterSpellChimeraShot | HunterSpellExplosiveShot |
		HunterSpellExplosiveTrap | HunterSpellBlackArrow | HunterSpellMultiShot | HunterSpellAimedShot |
		HunterSpellSerpentSting | HunterSpellKillShot | HunterSpellRapidFire | HunterSpellBestialWrath
)

// Agent is a generic way to access underlying hunter on any of the agents.
type HunterAgent interface {
	GetHunter() *Hunter
}
