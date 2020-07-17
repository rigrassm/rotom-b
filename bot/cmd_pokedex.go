package bot

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (b *Bot) handlePokedexCmd(
	s *discordgo.Session,
	env *commandEnvironment,
	m *discordgo.Message,
) error {

	if len(env.args) == 0 {
		return botError{
			title:   "Validation Error",
			details: "Please enter a Pokémon name to get its Pokédex info.",
		}
	}

	isShiny := strings.HasSuffix(env.args[0], "*") || strings.HasPrefix(env.args[0], "*")
	cleanPkmName := strings.ReplaceAll(env.args[0], "*", "")

	pkm, err := b.pokemonRepo.pokemon(strings.ToLower(cleanPkmName))
	if err != nil {
		return botError{
			title: "Pokémon not found",
			details: fmt.Sprintf("Pokémon %s could not be found.",
				cleanPkmName),
		}
	}

	externalPokedexLinks := fmt.Sprintf(
		"[Bulbapedia Entry](https://bulbapedia.bulbagarden.net/wiki/%s_(Pokémon))\n",
		strings.Title(pkm.Name),
	)
	if len(pkm.Dens.Shield) > 0 || len(pkm.Dens.Sword) > 0 || pkm.Generation == "SwordShield" {
		externalPokedexLinks += fmt.Sprintf(
			"[Serebii Sword & Shield Pokédex](https://serebii.net/pokedex-swsh/%s/)",
			strings.ToLower(pkm.Name),
		)
	} else {
		externalPokedexLinks += fmt.Sprintf(
			"[Serebii Sun & Moon Pokédex](https://serebii.net/pokedex-sm/%03d.shtml)",
			pkm.DexID,
		)
	}

	abilities := "`" + pkm.Abilities.Ability1
	if pkm.Abilities.Ability2 != "" {
		abilities += ",\n" + pkm.Abilities.Ability2
	}
	if pkm.Abilities.AbilityH != "" {
		abilities += ",\n" + pkm.Abilities.AbilityH + " (HA)"
	}
	abilities += "`"

	eggGroups := pkm.EggGroup1
	if pkm.EggGroup2 != "" {
		eggGroups = fmt.Sprintf(
			"%s, %s",
			pkm.EggGroup1,
			pkm.EggGroup2,
		)
	}

	forms := createJoinedPkmInfo("Forms", pkm.Forms)
	densSword := createJoinedPkmInfo("Sword", pkm.Dens.Sword)
	densShield := createJoinedPkmInfo("Shield", pkm.Dens.Shield)

	pkmForm := ""
	if len(env.args) > 1 {
		pkmForm = getSpriteForm(env.args[1])
	}

	embed := b.newEmbed()
	embed.Title = fmt.Sprintf("%s Pokédex Info", strings.Title(cleanPkmName))
	embed.Image = &discordgo.MessageEmbedImage{
		URL:    pkm.spriteImage(isShiny, pkmForm),
		Width:  300,
		Height: 300,
	}

	embed.URL = fmt.Sprintf(
		"https://bulbapedia.bulbagarden.net/wiki/%s_(Pokémon)",
		strings.ToLower(pkm.Name),
	)

	embed.Fields = []*discordgo.MessageEmbedField{
		&discordgo.MessageEmbedField{
			Name: "Base Stats",
			Value: fmt.Sprintf(
				"HP: `%d`\n"+
					"Atk: `%d`\n"+
					"Def: `%d`\n"+
					"Spa: `%d`\n"+
					"SpD: `%d`\n"+
					"Spe: `%d`\n"+
					"Total: `%d`",
				pkm.BaseStats.HP,
				pkm.BaseStats.Atk,
				pkm.BaseStats.Def,
				pkm.BaseStats.SpA,
				pkm.BaseStats.SpD,
				pkm.BaseStats.Spd,
				pkm.BaseStats.Total,
			),
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name:   "Abilities",
			Value:  abilities,
			Inline: true,
		},
		&discordgo.MessageEmbedField{
			Name: "Pokémon Misc. Info",
			Value: fmt.Sprintf(
				"Gender Ratio: `%s`\n"+
					"Height / Weight: `%s / %s`\n"+
					"Catch Rate: `%d`\n"+
					"Generation: `%s`\n"+
					"Egg Groups: `%s`\n"+
					"%s",
				pkm.GenderRatio,
				fmt.Sprintf("%.2f", pkm.Height),
				fmt.Sprintf("%.2f", pkm.Weight),
				pkm.CatchRate,
				pkm.Generation,
				eggGroups,
				forms,
			),
			Inline: true,
		},
	}

	if len(pkm.Dens.Shield) > 0 || len(pkm.Dens.Sword) > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name: "Dens",
			Value: fmt.Sprintf(
				"%s\n%s",
				densSword,
				densShield,
			),
			Inline: true,
		})
	}

	embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
		Name:   "More Info",
		Value:  externalPokedexLinks,
		Inline: false,
	})

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embed)
	return err
}

func createJoinedPkmInfo(prefix string, info []string) string {
	joinedInfo := ""
	if len(info) > 0 {
		joinedInfo = prefix + ": `" + strings.Join(info, ", ") + "`"
	}
	return joinedInfo
}
