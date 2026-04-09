# create-sample.ps1
# Creates 50 sample notes on the theme "cat behavior analysis".
# Notes are written directly as .md files with realistic spread timestamps
# to avoid the minute-resolution ID collision in the current note.json config.
#
# Distribution: 20 fleeting, 20 permanent, 10 project (slug: catbehavior)

$root = $PSScriptRoot

function Write-Note($path, $content) {
  $dir = Split-Path $path
  if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir | Out-Null }
  # Use UTF-8 without BOM, LF line endings so frontmatter_parse.go is happy
  $utf8NoBom = New-Object System.Text.UTF8Encoding $false
  [System.IO.File]::WriteAllText($path, $content.Replace("`r`n", "`n"), $utf8NoBom)
  Write-Host "  $path"
}

function Get-FrontmatterTitle($content) {
  $normalized = $content.Replace("`r`n", "`n")
  if (-not $normalized.StartsWith("---`n")) {
    return ""
  }
  $end = $normalized.IndexOf("`n---`n")
  if ($end -lt 0) {
    return ""
  }

  $header = $normalized.Substring(4, $end - 4)
  foreach ($line in ($header -split "`n")) {
    if ($line -match '^title:\s*(.+)$') {
      return $matches[1].Trim()
    }
  }
  return ""
}

function Add-MarkdownLinksForNoteType($rootPath, $relativeDir) {
  $utf8NoBom = New-Object System.Text.UTF8Encoding $false
  $targetDir = Join-Path $rootPath $relativeDir
  if (-not (Test-Path $targetDir)) {
    return
  }

  $files = Get-ChildItem -Path $targetDir -Filter *.md | Sort-Object Name
  if ($files.Count -lt 2) {
    return
  }

  $notes = @()
  foreach ($file in $files) {
    $content = [System.IO.File]::ReadAllText($file.FullName)
    $title = Get-FrontmatterTitle $content
    if ([string]::IsNullOrWhiteSpace($title)) {
      $title = [System.IO.Path]::GetFileNameWithoutExtension($file.Name)
    }
    $notes += [PSCustomObject]@{
      Path  = $file.FullName
      Name  = $file.Name
      Title = $title
    }
  }

  for ($i = 0; $i -lt $notes.Count; $i++) {
    $current = $notes[$i]
    $targets = @()
    for ($offset = 1; $offset -le 4; $offset++) {
      $targetIndex = ($i + $offset) % $notes.Count
      $targets += $notes[$targetIndex]
    }

    $linkLines = @()
    foreach ($target in $targets) {
      $linkLines += "- [$($target.Title)](./$($target.Name))"
    }
    $linksBlock = "## Links`n" + ($linkLines -join "`n")

    $content = [System.IO.File]::ReadAllText($current.Path).Replace("`r`n", "`n")
    $content = [System.Text.RegularExpressions.Regex]::Replace($content, "`n## Links`n[\s\S]*$", "")
    $content = $content.TrimEnd("`n") + "`n`n" + $linksBlock + "`n"
    [System.IO.File]::WriteAllText($current.Path, $content, $utf8NoBom)
  }
}

# ---------------------------------------------------------------------------
# FLEETING NOTES  (20)
# ---------------------------------------------------------------------------

Write-Note "$root\notes\fleeting\202601051823-first-outdoor-encounter.md" @"
---
id: 202601051823
title: First outdoor encounter observation
type: fleeting
created: 2026-01-05T18:23:00+01:00
tags: [outdoor, observation]
source:
---

Watched a cat approach a stranger cautiously — body low, tail down, nose forward.
Sniffed shoes then retreated to a safe distance. Classic risk assessment posture.
"@

Write-Note "$root\notes\fleeting\202601071042-tail-position-meanings.md" @"
---
id: 202601071042
title: Tail position meanings quick note
type: fleeting
created: 2026-01-07T10:42:00+01:00
tags: [body-language, tail]
source:
---

- High vertical tail = confident greeting
- Puffed tail = fear or aggression
- Tucked tail = submission or anxiety
- Slow swish = mild irritation
- Rapid flick = agitation
"@

Write-Note "$root\notes\fleeting\202601091534-kneading-behavior-observation.md" @"
---
id: 202601091534
title: Kneading behavior observation
type: fleeting
created: 2026-01-09T15:34:00+01:00
tags: [comfort, kneading]
source:
---

Cat kneads soft blanket before settling. Linked to nursing behavior retained from kittenhood.
Suggests comfort and security. Also activates scent glands in paw pads — subtle territory marking?
"@

Write-Note "$root\notes\fleeting\202601121856-grooming-frequency-note.md" @"
---
id: 202601121856
title: Grooming frequency note
type: fleeting
created: 2026-01-12T18:56:00+01:00
tags: [grooming, hygiene]
source:
---

Cats spend 30–50% of waking hours grooming. Excessive grooming can indicate stress (psychogenic alopecia).
Under-grooming often signals pain or illness. Worth tracking baseline for a specific individual.
"@

Write-Note "$root\notes\fleeting\202601151320-cat-eye-communication.md" @"
---
id: 202601151320
title: Cat eye communication
type: fleeting
created: 2026-01-15T13:20:00+01:00
tags: [eyes, communication]
source:
---

Dilated pupils in low light = normal. Dilated in bright light = fear, excitement, or pain.
Slow blink toward human = trust signal. Hard stare = threat. Half-closed = relaxed.
"@

Write-Note "$root\notes\fleeting\202601181647-hunting-instinct-in-play.md" @"
---
id: 202601181647
title: Hunting instinct visible in play
type: fleeting
created: 2026-01-18T16:47:00+01:00
tags: [play, hunting]
source:
---

Play mimics the full predatory sequence: stalk, pounce, grab, bite. Even sated cats play-hunt.
The motivational drive is separate from hunger — prey drive vs appetite.
"@

Write-Note "$root\notes\fleeting\202601211903-slow-blink-response.md" @"
---
id: 202601211903
title: Slow blink response experiment
type: fleeting
created: 2026-01-21T19:03:00+01:00
tags: [eyes, bonding]
source:
---

Tried slow-blinking at a shelter cat. It returned the slow blink twice within 30 seconds.
Kunstaetter 2020 study confirms: cats slow-blink more toward humans who slow-blink at them.
"@

Write-Note "$root\notes\fleeting\202601241235-vocalization-types.md" @"
---
id: 202601241235
title: Vocalization types quick list
type: fleeting
created: 2026-01-24T12:35:00+01:00
tags: [vocalization, communication]
source:
---

- Meow: largely directed at humans, rarely at other cats
- Trill/chirp: friendly greeting
- Chatter: hunting frustration at unreachable prey
- Hiss: warning
- Yowl: distress, mating, or territorial
- Purr: usually comfort but also stress/pain
"@

Write-Note "$root\notes\fleeting\202601271548-territory-marking-seen.md" @"
---
id: 202601271548
title: Territory marking methods seen
type: fleeting
created: 2026-01-27T15:48:00+01:00
tags: [territory, scent]
source:
---

Observed cat rubbing chin on corner of table, then cheek along wall.
Bunting deposits facial pheromones — marks familiar objects as safe.
Distinct from urine spraying which communicates to rival cats.
"@

Write-Note "$root\notes\fleeting\202602011022-window-watching-behavior.md" @"
---
id: 202602011022
title: Window watching behavior
type: fleeting
created: 2026-02-01T10:22:00+01:00
tags: [enrichment, indoor]
source:
---

Indoor cat spends ~2 hrs/day at window. Watches birds, reacts with chattering.
Provides environmental stimulation. High window perches reduce stress indicators.
"@

Write-Note "$root\notes\fleeting\202602041438-social-hierarchy-clue.md" @"
---
id: 202602041438
title: Social hierarchy clue in multi-cat home
type: fleeting
created: 2026-02-04T14:38:00+01:00
tags: [social, hierarchy]
source:
---

Dominant cat controls access to high perches and eats first. Subordinate waits.
No overt aggression — hierarchy maintained through avoidance and deference.
"@

Write-Note "$root\notes\fleeting\202602071819-nocturnal-activity-spike.md" @"
---
id: 202602071819
title: Nocturnal activity spike at 3am
type: fleeting
created: 2026-02-07T18:19:00+01:00
tags: [sleep, activity, crepuscular]
source:
---

Cat most active at dawn and dusk (crepuscular). Indoor cats shift patterns to owners' schedules partially.
The 3am zoomies = peak hunting-time energy release with nowhere to direct it.
"@

Write-Note "$root\notes\fleeting\202602101201-food-curiosity-vs-hunger.md" @"
---
id: 202602101201
title: Food curiosity vs hunger
type: fleeting
created: 2026-02-10T12:01:00+01:00
tags: [feeding, neophilia]
source:
---

Cats investigate new food even when not hungry — neophilia (novelty-seeking).
But neophobia also present: new food formats rejected if introduced abruptly.
Small portions of novel food alongside familiar food eases transition.
"@

Write-Note "$root\notes\fleeting\202602131647-scent-exchange-greeting.md" @"
---
id: 202602131647
title: Scent exchange in cat greeting
type: fleeting
created: 2026-02-13T16:47:00+01:00
tags: [greeting, scent, social]
source:
---

Two familiar cats meet: nose-to-nose first, then flank and tail-base sniff.
Exchanging individual odor profiles ("allogrooming with nose").
Strange cats skip nose contact and go straight to tail-base or not at all.
"@

Write-Note "$root\notes\fleeting\202602161023-stress-signals-checklist.md" @"
---
id: 202602161023
title: Stress signals checklist
type: fleeting
created: 2026-02-16T10:23:00+01:00
tags: [stress, welfare]
source:
---

- Hiding more than usual
- Reduced grooming or over-grooming
- Loss of appetite
- Inappropriate elimination
- Increased vocalization
- Aggression toward familiar people
- Freezing/crouching in open space
"@

Write-Note "$root\notes\fleeting\202602191534-comfort-seeking-in-storms.md" @"
---
id: 202602191534
title: Comfort seeking during thunder
type: fleeting
created: 2026-02-19T15:34:00+01:00
tags: [stress, fear, weather]
source:
---

Cat sought narrow enclosed space under bed during thunderstorm. Classic den-seeking.
Did not want petting — touch escalates arousal when already stressed.
Presence nearby without contact seemed to reduce duration of hiding.
"@

Write-Note "$root\notes\fleeting\202602221847-attention-seeking-tactics.md" @"
---
id: 202602221847
title: Attention-seeking tactics catalog
type: fleeting
created: 2026-02-22T18:47:00+01:00
tags: [attention, social, vocalization]
source:
---

- Sitting on keyboard / book
- Knocking objects off tables
- Head-butting or pawing face
- Persistent meowing escalating in pitch
- Bringing prey (real or toy) as gift

Operant conditioning: these behaviors were reinforced by owner attention at some point.
"@

Write-Note "$root\notes\fleeting\202603011209-carrier-aversion-cause.md" @"
---
id: 202603011209
title: Carrier aversion root causes
type: fleeting
created: 2026-03-01T12:09:00+01:00
tags: [fear, carrier, conditioning]
source:
---

Carrier = only appears before vet visit. Negative association formed quickly for context-sensitive animals.
Counter-conditioning: leave carrier out with bedding, feed meals inside. Takes 4–8 weeks.
"@

Write-Note "$root\notes\fleeting\202603041523-stranger-assessment-process.md" @"
---
id: 202603041523
title: How cats assess strangers
type: fleeting
created: 2026-03-04T15:23:00+01:00
tags: [social, strangers, assessment]
source:
---

Cats approach strangers who ignore them faster than those who reach toward them.
Optimal greeting: sit at floor level, extend one finger at nose height, let cat approach.
Avoid direct eye contact initially — reads as threatening.
"@

Write-Note "$root\notes\fleeting\202603071836-playful-aggression-signs.md" @"
---
id: 202603071836
title: Playful vs real aggression signs
type: fleeting
created: 2026-03-07T18:36:00+01:00
tags: [aggression, play]
source:
---

Playful: claws retracted or gentle, no vocalizing, rebounds immediately, pupils dilated.
Real: hissing, growling, ears flat, full claw contact, prolonged freeze before attack.
Misdirected play aggression in young cats often mistaken for real aggression.
"@

# ---------------------------------------------------------------------------
# PERMANENT NOTES  (20)
# ---------------------------------------------------------------------------

Write-Note "$root\notes\permanent\202601061445-cat-communication-overview.md" @"
---
id: 202601061445
title: Cat communication overview
type: permanent
created: 2026-01-06T14:45:00+01:00
tags: [communication, reference]
updated: 2026-02-15T09:00:00+01:00
---

Cats communicate through four channels simultaneously:

**Vocal:** meow, trill, chatter, hiss, purr, yowl. Most vocalizations directed at humans, not other cats.

**Visual:** body posture, tail position, ear orientation, eye shape, whisker position.

**Olfactory:** facial pheromones (bunting), urine/feces marking, anal gland secretions, paw gland deposits.

**Tactile:** allogrooming, head-butting (bunting), kneading, play contact.

Reading cats requires integrating all four channels. A purring cat with flattened ears is not simply content.
"@

Write-Note "$root\notes\permanent\202601091902-feline-social-structure.md" @"
---
id: 202601091902
title: Feline social structure
type: permanent
created: 2026-01-09T19:02:00+01:00
tags: [social, ecology]
updated: 2026-02-20T11:00:00+01:00
---

Domestic cats are facultatively social — capable of both solitary and group living depending on resource distribution.

**Feral colonies** form around concentrated food sources. Females related by kinship form the stable core; males are peripheral and transient. Colony members share kitten-rearing (communal nursing).

**Pair/group formation** in domestic settings is driven by early socialization. Littermates raised together form stronger bonds than adult-introduced cats.

**Implications for multi-cat households:** unrelated adult cats introduced without gradual territory negotiation default to avoidance hierarchies, not friendship.
"@

Write-Note "$root\notes\permanent\202601131523-predatory-play-patterns.md" @"
---
id: 202601131523
title: Predatory play patterns in domestic cats
type: permanent
created: 2026-01-13T15:23:00+01:00
tags: [play, hunting, ethology]
updated: 2026-03-01T10:00:00+01:00
---

Play in cats is not juvenile behavior that disappears with age. It is a persistent expression of predatory motivation decoupled from hunger.

**Behavioral sequence (after Leyhausen 1979):**
1. Orient
2. Stalk (slow, body low, weight forward)
3. Rush
4. Pounce/Bat
5. Grab-bite (neck bite in hunting; inhibited in play)
6. Dissect (object manipulation)

Cats cycle through this sequence even with inanimate objects. Sequence can loop at any stage.

**Play deprivation** increases redirected aggression and stereotypic behaviors. 2×10-minute interactive play sessions per day significantly reduce problem behaviors.
"@

Write-Note "$root\notes\permanent\202601161047-scent-marking-taxonomy.md" @"
---
id: 202601161047
title: Scent marking taxonomy
type: permanent
created: 2026-01-16T10:47:00+01:00
tags: [scent, territory, behavior]
updated: 2026-02-10T14:00:00+01:00
---

Cats deposit scent through several distinct mechanisms with different social functions:

| Method | Glands | Function |
|---|---|---|
| Bunting (head rub) | Facial, chin, cheeks | Marks safe/familiar objects and individuals |
| Flank rub | Flank glands | Territory boundary marking |
| Kneading | Interdigital paw glands | Comfort + subtle surface marking |
| Urine spray | Urine + anal secretions | Active communication to rivals; stress marker |
| Middening (feces exposure) | Anal glands | Strong territory assertion, rare in confident cats |

Bunting is socially positive. Urine spraying in an established indoor cat almost always signals environmental stress.
"@

Write-Note "$root\notes\permanent\202601191834-sleep-cycles-and-activity.md" @"
---
id: 202601191834
title: Sleep cycles and activity windows
type: permanent
created: 2026-01-19T18:34:00+01:00
tags: [sleep, activity, physiology]
updated: 2026-02-25T09:30:00+01:00
---

Cats sleep 12–16 hours per day. Most sleep is light (dozing); deep REM sleep accounts for ~25%.

**Crepuscular baseline:** genetically primed for dawn/dusk activity matching prey activity cycles in the wild.

**Domestic adjustment:** indoor cats partially synchronize to owner schedules through social learning and meal-time conditioning, but the 3–5am activity spike persists without sufficient daytime stimulation.

**REM behavior:** muscle twitching, whisker movement, and vocalizations during sleep are normal dream-state activity — homologous to human REM.

Enriching waking hours (play, puzzle feeders, outdoor access) reduces night disturbance.
"@

Write-Note "$root\notes\permanent\202601221256-territorial-behavior-analysis.md" @"
---
id: 202601221256
title: Territorial behavior analysis
type: permanent
created: 2026-01-22T12:56:00+01:00
tags: [territory, aggression, space]
updated: 2026-03-05T11:00:00+01:00
---

Territory in cats is not a fixed defended area but a fluid set of time-shared spaces.

**Core area:** sleeping and feeding zone; most actively defended; rarely shared with non-bonded cats.

**Home range:** broader patrol zone; overlaps with neighboring cats via time-sharing (different cats use the same path at different hours).

**Time-sharing** reduces direct conflict. Disruption (new cat, construction, rearranged furniture) forces renegotiation and elevates cortisol.

**Indoor vs outdoor:** indoor-only cats have a single compressed territory. Resource distribution within the home (feeding stations, litter boxes, resting spots) determines quality of life. Rule: N+1 resources per cat.
"@

Write-Note "$root\notes\permanent\202601251619-body-language-reference.md" @"
---
id: 202601251619
title: Body language reference guide
type: permanent
created: 2026-01-25T16:19:00+01:00
tags: [body-language, reference]
updated: 2026-03-10T08:00:00+01:00
---

**Ears:**
- Forward = alert, interested
- Rotated sideways = mild irritation
- Flat back = fear or aggression

**Tail:**
- Vertical = confident greeting
- Curved at tip = playful interest
- Puffed = fear/threat
- Low and tucked = submission

**Eyes:**
- Slow blink = relaxed, trust
- Hard stare = challenge
- Dilated pupils (bright light) = arousal (fear, excitement, pain)

**Overall posture:**
- Side-on, arched back, fur raised = fear threat display (appears larger)
- Crouched, head down, ears flat = pure fear
- Roll over exposing belly = trust (NOT always an invitation to touch)

The belly is often misread. Exposed belly = high trust display. Grabbing it triggers defensive reflex.
"@

Write-Note "$root\notes\permanent\202601281034-vocalizations-classification.md" @"
---
id: 202601281034
title: Vocalizations classification
type: permanent
created: 2026-01-28T10:34:00+01:00
tags: [vocalization, communication]
updated: 2026-02-18T15:00:00+01:00
---

Moelk (1944) classified cat vocalizations into three groups by mouth position:

**Closed mouth (murmurs):** purr, trill, chirrup — produced continuously, used in friendly/affiliative contexts.

**Open-then-closing mouth (vowel sounds):** meow and variants — highly variable; adult cats rarely use with each other, learned for human communication.

**Fixed open mouth (intense sounds):** hiss, spit, growl, yowl, snarl, chatter — agonistic and mating contexts.

**Meow plasticity:** individual cats develop custom meow repertoires shaped by owner responses. A cat's meow register differs meaningfully between households.
"@

Write-Note "$root\notes\permanent\202602021548-attachment-styles-in-cats.md" @"
---
id: 202602021548
title: Attachment styles in domestic cats
type: permanent
created: 2026-02-02T15:48:00+01:00
tags: [attachment, bonding, psychology]
updated: 2026-03-12T10:00:00+01:00
---

Vitale et al. (2019) applied the Strange Situation Test to cats. Results showed attachment style distribution:

- **Secure (65%):** uses owner as safe base, explores, mild distress on separation, settles on return.
- **Ambivalent (11%):** high distress, clingy on return, slow to settle.
- **Avoidant (11%):** little distress, ignores owner on return.
- **Disorganized (13%):** inconsistent, erratic on return.

Kitten socialization window (2–7 weeks) critically shapes adult attachment capacity. Early positive human contact produces secure attachment regardless of breed.

Attachment style is stable across stress conditions — it is a trait, not a state.
"@

Write-Note "$root\notes\permanent\202602051923-stress-and-environment.md" @"
---
id: 202602051923
title: Stress and environment in cats
type: permanent
created: 2026-02-05T19:23:00+01:00
tags: [stress, welfare, environment]
updated: 2026-03-08T09:00:00+01:00
---

Environmental stress in cats is under-diagnosed because most stress responses are behavioral and gradual.

**Primary stressors:**
- Unpredictable schedule (feeding, human presence)
- Insufficient vertical space
- Litter box location, number, or cleanliness
- Other cats (conflict or competition)
- Loud / unpredictable sounds
- Inability to hide

**Secondary markers (often misattributed):**
- Idiopathic cystitis — strongly stress-correlated
- Over-grooming / barbering
- Inappropriate elimination
- Reduced play engagement

MEMO environmental checklist: hiding spots, elevated perches, predictable routine, N+1 resources.
"@

Write-Note "$root\notes\permanent\202602081307-hunting-sequence-model.md" @"
---
id: 202602081307
title: Hunting sequence model
type: permanent
created: 2026-02-08T13:07:00+01:00
tags: [hunting, ethology, model]
updated: 2026-02-28T11:00:00+01:00
---

The complete hunting sequence is rarely executed in domestic settings. Cats typically abort at pounce or grab.

**Full sequence:** search → orient → stalk → rush → pounce → grab-bite → kill-bite → consume

**Key insight (Biben 1979):** hunger increases the probability of completing the sequence to consumption, but does NOT initiate play-hunting. A satiated cat will hunt and play; a starving cat may not play but will hunt to consume.

**Implications:**
- Interactive wand play should engage all sequence stages to be satisfying — not just batting
- "Fishing rod" toys that allow full stalk → pounce → grab are more enriching than ping-pong balls
- Allow completion — ending play abruptly at pounce leaves the sequence incomplete and may increase frustration
"@

Write-Note "$root\notes\permanent\202602111742-grooming-social-function.md" @"
---
id: 202602111742
title: Grooming as social function
type: permanent
created: 2026-02-11T17:42:00+01:00
tags: [grooming, social, allogrooming]
updated: 2026-03-03T10:00:00+01:00
---

Individual grooming (autogrooming) serves hygiene and thermoregulation. Social grooming (allogrooming) has additional functions:

**Allogrooming targets:** head and neck (areas cat cannot self-groom). Always initiated by the groomer, not the receiver — it is an active affiliation signal.

**Social bonding:** allogrooming is asymmetric in multi-cat groups. Dominant cats give AND receive. Subordinates receive less. Bonded pairs groom reciprocally.

**Stress response:** some cats over-groom when anxious. Distinguishing psychogenic alopecia from physical dermatitis requires behavioral history alongside dermatological exam.

**Human-cat:** owner petting the head/neck region = proxy allogrooming. Cats that head-butt human hands are soliciting this.
"@

Write-Note "$root\notes\permanent\202602141056-human-cat-bond-formation.md" @"
---
id: 202602141056
title: Human-cat bond formation
type: permanent
created: 2026-02-14T10:56:00+01:00
tags: [bonding, human-animal, socialization]
updated: 2026-03-07T14:00:00+01:00
---

Bond formation is bidirectional and time-sensitive.

**Critical period (kittens):** 2–7 weeks. Exposure to humans, handling, sounds, smells during this window produces confident, sociable adult cats. Miss this window and adult socialization is possible but requires far greater effort.

**Adult bond formation:** requires cat to initiate contact on its own terms. Owner-initiated contact that the cat tolerates is not the same as affiliative bonding. Let the cat lead.

**Bond quality indicators:**
- Cat seeks proximity without food motivation
- Cat sleeps in contact with owner
- Cat slow-blinks, trills, or allogrooms owner
- Cat returns to owner after startle events

**Misconception:** independence = lack of attachment. Secure cats are confident enough to explore. They still use the owner as a safe base.
"@

Write-Note "$root\notes\permanent\202602171523-fear-response-triggers.md" @"
---
id: 202602171523
title: Fear response triggers and patterns
type: permanent
created: 2026-02-17T15:23:00+01:00
tags: [fear, behavior, stress]
updated: 2026-03-09T09:00:00+01:00
---

Fear responses in cats are mediated by the amygdala and follow three patterns:

**Freeze:** first response; reduces detection by predators. Cat becomes immobile, pupils dilate, breathing shallow.

**Flight:** preferred when escape route available. Cat retreats to elevated or enclosed safe zone.

**Fight:** last resort when cornered. Includes hissing, spitting, scratching, biting. Often preceded by warning displays.

**Classic triggers:**
- Novel objects or environments
- Sudden movement or sound
- Unfamiliar animals or people
- Veterinary contexts (conditioned fear)
- Loss of predictability in routine

**Treatment:** systematic desensitization + counter-conditioning. Never flood (forced exposure) — reliably worsens fear.
"@

Write-Note "$root\notes\permanent\202602201847-play-deprivation-effects.md" @"
---
id: 202602201847
title: Effects of play deprivation
type: permanent
created: 2026-02-20T18:47:00+01:00
tags: [play, welfare, enrichment]
updated: 2026-03-11T10:00:00+01:00
---

Play deprivation is a welfare concern rarely taken as seriously as nutritional or medical needs.

**Documented effects:**
- Increased redirected aggression (toward objects, other pets, owners)
- Stereotypic behaviors (repetitive pacing, over-vocalization)
- Increased anxiety and hiding
- Obesity (reduced activity expenditure)
- Impaired learning and problem-solving

**Minimum recommendation:** 2 interactive play sessions of 10–15 minutes daily, using wand/feather toys that engage full predatory sequence. Puzzle feeders replace some play function for indoor cats.

**Play in senior cats:** often reduced due to arthritis, not disinterest. Gentler, shorter sessions with less jumping remain beneficial and should be maintained.
"@

Write-Note "$root\notes\permanent\202602231234-multi-cat-household-dynamics.md" @"
---
id: 202602231234
title: Multi-cat household dynamics
type: permanent
created: 2026-02-23T12:34:00+01:00
tags: [social, multi-cat, conflict]
updated: 2026-03-14T11:00:00+01:00
---

Cats are not obligate social animals. Multi-cat households work best with deliberate management.

**Introduction protocol:**
1. Separate rooms with scent exchange (swap bedding) for 1–2 weeks
2. Visual access without physical contact (baby gate or door gap)
3. Supervised free access with escape routes for both cats
4. Fully integrated when no sustained vigilance or avoidance observed

**Resource distribution:** N+1 rule. If 2 cats: 3 food stations, 3 water sources, 3 litter boxes, multiple resting spots at varied heights.

**Chronic conflict signals:** sustained staring, blocking access to resources, sudden attacks without warning, stress-related illness in either cat.

Inter-cat aggression that doesn't resolve with resources/space may require pheromone therapy (Feliway Multicat) or veterinary behaviorist consultation.
"@

Write-Note "$root\notes\permanent\202603021618-cognitive-enrichment-needs.md" @"
---
id: 202603021618
title: Cognitive enrichment needs
type: permanent
created: 2026-03-02T16:18:00+01:00
tags: [enrichment, cognition, welfare]
updated: 2026-03-15T09:00:00+01:00
---

Cats have greater cognitive capacity than their reputation suggests. Boredom is a real welfare issue.

**Documented problem-solving:** cats learn by observation (social learning), trial-and-error, and can recall solutions after delays of 24 hours.

**Enrichment categories:**
- **Sensory:** new scents, varied textures, window birdfeeders
- **Food:** puzzle feeders, scatter feeding, lick mats (slows eating, extends engagement)
- **Social:** interactive play with owner or compatible companion
- **Environmental:** rotating toys (familiarity reduces engagement within ~3 days)
- **Training:** cats can be clicker-trained; desensitizes vet handling, provides mental stimulation

**Key principle:** novelty drives engagement. Rotating enrichment items on a 3-day cycle maintains interest better than permanent availability.
"@

Write-Note "$root\notes\permanent\202603051043-pain-behavior-indicators.md" @"
---
id: 202603051043
title: Pain behavior indicators in cats
type: permanent
created: 2026-03-05T10:43:00+01:00
tags: [pain, health, behavior]
updated: 2026-03-16T08:00:00+01:00
---

Cats mask pain — evolutionary vestige of prey-animal self-protection. Subtle behavioral changes are often the only signal.

**Acute pain indicators:**
- Guarding affected body part
- Reduced grooming of specific area (or increased licking of painful site)
- Uncharacteristic aggression when touched
- Vocalizing when moving or handled

**Chronic pain indicators (e.g., arthritis):**
- Reduced jumping height or avoidance of heights
- Hesitation at stairs
- Reduced play engagement
- Increased sleep, reduced exploration
- Facial grimace scale (Feline Grimace Scale, validated 2019): orbital tightening, muzzle tension, ear position, whisker position, head position

**Action threshold:** any behavior change lasting >48 hours warrants investigation. Pain is often undertreated in cats due to limited owner recognition.
"@

Write-Note "$root\notes\permanent\202603081757-indoor-vs-outdoor-behavior.md" @"
---
id: 202603081757
title: Indoor vs outdoor behavioral differences
type: permanent
created: 2026-03-08T17:57:00+01:00
tags: [indoor, outdoor, welfare]
updated: 2026-03-17T10:00:00+01:00
---

Indoor and outdoor cats differ in stress profiles, behavioral repertoire, and risk exposure.

**Outdoor advantages:** full predatory sequence expression, territorial range, environmental novelty, vitamin D synthesis, richer sensory environment.

**Indoor advantages:** no road injury, predation, disease transmission from wildlife; measurably longer average lifespan (10–15 yrs indoor vs 2–5 yrs outdoor in urban environments).

**Behavioral compensations for indoor cats:**
- Window perches with bird/squirrel feeders outside
- Regular interactive play covering full predatory sequence
- Outdoor enclosures ("catios") as compromise
- Leash walks for confident cats — requires early training

**Behavioral problems more common indoors:** obesity, boredom-related stereotypy, inter-cat conflict (compressed territory), idiopathic cystitis.
"@

Write-Note "$root\notes\permanent\202603111325-aging-and-behavior-change.md" @"
---
id: 202603111325
title: Aging and behavior change in cats
type: permanent
created: 2026-03-11T13:25:00+01:00
tags: [aging, senior, behavior]
updated: 2026-03-18T09:00:00+01:00
---

Cats are considered senior at 11+ years, geriatric at 15+.

**Normal age-related changes:**
- Reduced activity and play interest
- Increased sleep duration
- Decreased sensory acuity (hearing/vision)
- Slowed thermoregulation (seeks warmth more)
- Reduced social solicitation

**Pathological changes (warrant investigation):**
- Cognitive Dysfunction Syndrome (CDS): disorientation, altered sleep-wake cycle, vocalization at night, loss of litter training
- Hyperthyroidism: increased vocalization, weight loss despite appetite, hyperactivity
- Hypertension: sudden behavioral change, apparent vision loss, changed temperament
- Chronic kidney disease: increased water intake, reduced appetite, weight loss

**CDS vs normal aging:** key marker is confusion in familiar environments. Night-time yowling in senior cats is CDS until proven otherwise.
"@

# ---------------------------------------------------------------------------
# PROJECT NOTES  (10)  — project slug: catbehavior
# ---------------------------------------------------------------------------

Write-Note "$root\notes\project\202601041230-catbehavior-project-overview.md" @"
---
id: 202601041230
title: Cat behavior analysis project overview
type: project
project: catbehavior
created: 2026-01-04T12:30:00+01:00
tags: [project, overview]
updated: 2026-03-19T10:00:00+01:00
---

## Goal

Build a personal reference knowledge base on domestic cat behavior covering:
- Communication systems (vocal, visual, olfactory, tactile)
- Social organization and attachment
- Predatory and play behavior
- Welfare, stress, and enrichment
- Applied behavioral observations case studies

## Deliverables

- 20 permanent notes forming reference knowledge base
- 2 case study notes (breed-specific behavioral profiles)
- Synthesis note integrating main findings

## Timeline

Jan–Mar 2026: literature review and observation notes
Apr 2026: synthesis and finalization
"@

Write-Note "$root\notes\project\202601081547-catbehavior-observation-methodology.md" @"
---
id: 202601081547
title: Observation methodology for cat behavior
type: project
project: catbehavior
created: 2026-01-08T15:47:00+01:00
tags: [methodology, observation]
updated: 2026-02-05T11:00:00+01:00
---

## Sampling Method

**Focal animal sampling:** observe one individual for a fixed period (10 min), record all behaviors.
**Scan sampling:** at fixed intervals (every 5 min), record current behavior of all animals present.

## Behavior Categories

| Code | Behavior |
|---|---|
| GR | Grooming (self) |
| AG | Allogrooming |
| PL | Play |
| EX | Explore |
| RE | Rest/Sleep |
| EA | Eat/Drink |
| EL | Eliminate |
| VL | Vocalize |
| HI | Hide |
| AL | Alert/Watch |

## Recording

Video capture preferred for playback analysis. Note environmental context for each session.
Baseline: min 3 sessions per individual before drawing behavioral conclusions.
"@

Write-Note "$root\notes\project\202601151903-catbehavior-literature-review.md" @"
---
id: 202601151903
title: Literature review notes
type: project
project: catbehavior
created: 2026-01-15T19:03:00+01:00
tags: [literature, references]
updated: 2026-02-22T10:00:00+01:00
---

## Key Sources

**Leyhausen, P. (1979)** — Cat Behavior. Foundational ethological account of predatory and social behavior.

**Turner, D.C. & Bateson, P. (2000)** — The Domestic Cat: The Biology of its Behaviour. Standard reference text.

**Vitale, K.R. et al. (2019)** — "Attachment bonds between domestic cats and humans." Current Biology. Secure attachment study.

**Moelk, M. (1944)** — "Vocalizing in the house-cat." American Journal of Psychology. Vocalization taxonomy.

**Biben, M. (1979)** — "Predation and predatory play." Animal Behaviour. Hunger vs. play motivation.

**Feline Grimace Scale (2019)** — Evangelista et al. Validated acute pain assessment tool.

## Gaps in Literature

- Long-term attachment stability studies
- Behavioral differences across urban density levels
- Effect of owner personality on cat stress levels
"@

Write-Note "$root\notes\project\202601221138-catbehavior-behavioral-categories.md" @"
---
id: 202601221138
title: Behavioral categories framework
type: project
project: catbehavior
created: 2026-01-22T11:38:00+01:00
tags: [framework, categories]
updated: 2026-02-12T09:00:00+01:00
---

## Top-Level Behavioral Categories

1. **Maintenance behaviors** — grooming, sleeping, eating, elimination
2. **Exploratory behaviors** — sniffing, investigating, patrolling
3. **Predatory behaviors** — hunting sequence, play-hunting
4. **Social behaviors** — greeting, allogrooming, play, conflict
5. **Communication** — vocalizations, body posture, scent marking
6. **Stress/Fear responses** — hiding, freezing, flight, aggression
7. **Affiliative behaviors** — bunting, kneading, proximity-seeking

## Cross-cutting Themes

- Many behaviors serve multiple functions (kneading = comfort + marking)
- Context determines interpretation (growl in play vs threat)
- Individual variation is large; population-level claims need sample sizes
"@

Write-Note "$root\notes\project\202601291612-catbehavior-data-collection-plan.md" @"
---
id: 202601291612
title: Data collection plan
type: project
project: catbehavior
created: 2026-01-29T16:12:00+01:00
tags: [data, plan]
updated: 2026-02-08T10:00:00+01:00
---

## Subjects

- Primary: 2 owned domestic cats (neutered male DSH, spayed female tabby)
- Secondary: shelter observation sessions (Tuesdays, 2 hrs)

## Variables Tracked

| Variable | Method |
|---|---|
| Daily activity budget | Scan sampling, 5-min intervals, 1 hr morning + 1 hr evening |
| Social interaction frequency | Event sampling, all visible interactions |
| Vocalization contexts | Audio notes tagged with trigger context |
| Stress indicators | Daily checklist (grooming, appetite, elimination, hiding) |

## Data Format

Markdown notes (fleeting) per session, promoted to permanent when pattern emerges.
Tags: #observation, #shelter, #subject-a, #subject-b
"@

Write-Note "$root\notes\project\202602051047-catbehavior-midpoint-summary.md" @"
---
id: 202602051047
title: Midpoint project summary
type: project
project: catbehavior
created: 2026-02-05T10:47:00+01:00
tags: [summary, progress]
updated: 2026-02-28T11:00:00+01:00
---

## Progress at Week 5

**Completed:**
- Observation methodology established
- Literature review notes (6 key sources)
- 12 fleeting capture notes
- 8 permanent reference notes

**Key findings so far:**
- Communication is multimodal — single-channel reading systematically misleads
- Play and predation share a behavioral substrate but have different motivational sources
- Stress indicators are subtle and cumulative; daily baselines matter

**Adjustments:**
- Increase shelter observation from 1x to 2x per week
- Add enrichment experiment: puzzle feeder vs scatter feeding engagement duration

**Remaining:**
- Case study notes (Siamese, Maine Coon)
- Synthesis note
- Final review of all permanent notes for consistency
"@

Write-Note "$root\notes\project\202602121534-catbehavior-pattern-analysis.md" @"
---
id: 202602121534
title: Pattern analysis draft
type: project
project: catbehavior
created: 2026-02-12T15:34:00+01:00
tags: [analysis, patterns]
updated: 2026-03-10T09:00:00+01:00
---

## Emerging Patterns from Observation Data

### Activity Budget (Subject A — male, 4 yrs, indoor-only)
- Rest/sleep: 58% of observed time
- Grooming: 18%
- Alert/window: 12%
- Play-initiated: 6%
- Eating: 4%
- Explore: 2%

### Vocalization Triggers
- 72% of vocalizations preceded owner entering room (anticipatory)
- 18% during play (chirp/chatter)
- 10% at window (prey-directed chatter)

### Stress Checklist Trends
- Baseline stable across 4 weeks
- Single spike in week 3 correlating with workman visit (noise + strangers)
- Recovery to baseline within 48 hrs

## Preliminary Conclusions
- Activity budget aligns with literature (sedentary majority, peaks at crepuscular)
- High anticipatory vocalization suggests strong owner-association conditioning
- Stress resilience moderate-good: recovered quickly from acute stressor
"@

Write-Note "$root\notes\project\202602191922-catbehavior-case-study-siamese.md" @"
---
id: 202602191922
title: Case study — Siamese behavioral profile
type: project
project: catbehavior
created: 2026-02-19T19:22:00+01:00
tags: [case-study, siamese, breed]
updated: 2026-03-12T10:00:00+01:00
---

## Subject

Siamese female, 3 years, spayed. Indoor-only. Single-cat household. Owner works from home.

## Notable Behavioral Traits

**Vocalization:** Exceptionally high. Yowl-frequency at least 3x baseline DSH. Vocal in response to owner attention, food, doors opening. Yowl pitch escalates if unacknowledged.

**Social demand:** Follows owner room-to-room. Initiates contact every 15–20 min on average. Does not tolerate owner absence >4 hrs without stress indicators.

**Intelligence/problem-solving:** Opened lever-style door handle (observed twice). Learned puzzle feeder format within 2 sessions. Anticipates schedule events 10–15 min early.

**Body sensitivity:** Resists grooming of lower back and tail base. Allows full head/neck grooming. Consistent with typical Siamese tactile sensitivity.

## Behavioral Notes

High sociality and vocalization appear to be breed-typical, not individual pathology. Owner management strategy: scheduled interactive play twice daily reduces attention-solicitation frequency by ~40%.
"@

Write-Note "$root\notes\project\202602261348-catbehavior-case-study-maine-coon.md" @"
---
id: 202602261348
title: Case study — Maine Coon behavioral profile
type: project
project: catbehavior
created: 2026-02-26T13:48:00+01:00
tags: [case-study, maine-coon, breed]
updated: 2026-03-14T09:00:00+01:00
---

## Subject

Maine Coon male, 5 years, neutered. Indoor-outdoor. Two-cat household (bonded pair, littermates).

## Notable Behavioral Traits

**Size and movement:** 8.2kg. Movement noticeably deliberate and slow compared to DSH. High perching preference — occupies highest available point consistently.

**Vocalization:** Relatively quiet. Uses trill and chirrup for greetings. Meow rare; reserved for food solicitation only.

**Water interaction:** Fascinates with moving water. Paws at water bowl, drinks from running tap. Consistent with breed-typical trait (possibly linked to semi-aquatic working history).

**Social behavior:** Affiliative with bonded littermate — mutual grooming 2–3x daily. Tolerant of strangers after 5–10 min assessment. Does not hide during visitor presence after assessment phase.

**Play style:** Prefers object manipulation over chase. Bats and carries objects. Full predatory sequence to "kill-bite" stage on toys — more complete than observed in lighter breeds.

## Behavioral Notes

Maine Coon temperament well-suited to multi-cat environment. Slow arousal curve — stress indicators rarely spike. Water behavior should be accommodated (fountain feeder).
"@

Write-Note "$root\notes\project\202603051723-catbehavior-synthesis-outline.md" @"
---
id: 202603051723
title: Final synthesis outline
type: project
project: catbehavior
created: 2026-03-05T17:23:00+01:00
tags: [synthesis, final]
updated: 2026-03-20T09:00:00+01:00
---

## Synthesis Structure

### 1. The Misread Cat
Core thesis: domestic cats are systematically misunderstood because human observers default to single-channel reading and anthropomorphic interpretation.

### 2. Communication as Multimodal System
Summary of vocal + visual + olfactory + tactile simultaneously. Key: conflicting signals (purring + flat ears) are common and diagnostically important.

### 3. Social Flexibility vs Social Obligation
Cats are not loners and not pack animals. Facultative sociality means context determines behavior — the same cat can be affiliative and solitary depending on resource, familiarity, and early socialization.

### 4. The Predatory Core
Play = truncated hunting. This single insight explains most "problem behaviors": midnight zoomies, ankle attacks, chattering, grass-eating (digestive aid for indigestible prey parts).

### 5. Welfare as Behavioral Expression
Stress and pain manifest behaviorally, subtly. Owners who understand behavioral baselines catch medical problems earlier.

### 6. Individual Variation
Breed tendencies are real but weak predictors. Individual socialization history and owner behavior are stronger predictors of adult temperament than genetics in most cases.

## Next Steps

- Write draft synthesis note in permanent notes
- Cross-link all relevant permanent notes
- Share with study group for peer review
"@

Write-Host ""
Add-MarkdownLinksForNoteType $root "notes\permanent"
Add-MarkdownLinksForNoteType $root "notes\project"
Write-Host "Added markdown links across permanent and project notes."
Write-Host ""
Write-Host "Done. Created 50 notes:"
Write-Host "  20 fleeting  -> notes/fleeting/"
Write-Host "  20 permanent -> notes/permanent/"
Write-Host "  10 project   -> notes/project/  (project: catbehavior)"
