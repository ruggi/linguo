# Linguo

![](http://i.imgur.com/LYYiWx5.jpg)

Linguo is a Go [Natural Language Processing](https://en.wikipedia.org/wiki/Natural_language_processing) library.

Linguo is based on [go-freeling](https://github.com/advancedlogic/go-freeling), which is a port of [Freeling](http://nlp.lsi.upc.edu/freeling/). Linguo started as an attempt to clean up and improve go-freeling, adding more features on top of it.

This is still a WIP and more documentation will follow.

## Usage

`go get -u github.com/reifcode/linguo`

```
import "github.com/reifcode/linguo"

...

engine := linguo.NewEngine()
engine.InitNLP("./data", "en")
result := engine.NLP.Workflow("Linguo was a grammar-correcting robot created by Lisa Simpson.")
...
```

Note: Linguo uses [MITIE](https://github.com/mit-nlp/MITIE) for entity extraction, so be sure to have it installed. On MacOS you can just install it with Homebrew:

```
$ brew install mitie
```


TBC

## Examples

See `examples/example.go`.

```
$ go run examples/example.go
Sentence: "Linguo was a grammar-correcting robot created by Lisa_Simpson ."
Entities:
	* [NP] Linguo (linguo) 1.0000%
	* [VBD] was (be) 1.0000%
	* [DT] a (a) 0.5000%
	* [VBG] grammar-correcting (grammar-correcting) 1.0000%
	* [NN] robot (robot) 1.0000%
	* [VBN] created (create) 0.8027%
	* [IN] by (by) 0.9972%
	* [NP] Lisa_Simpson (lisa_simpson) 1.0000%
	* [Fp] . (.) 1.0000%

Sentence: "It is from the eighteenth episode of Season_12 ."
Entities:
	* [PRP] It (it) 1.0000%
	* [VBZ] is (be) 1.0000%
	* [IN] from (from) 1.0000%
	* [DT] the (the) 1.0000%
	* [JJ] eighteenth (18) 0.6002%
	* [NN] episode (episode) 1.0000%
	* [IN] of (of) 0.9999%
	* [NP] Season_12 (season_12) 1.0000%
	* [Fp] . (.) 1.0000%

Entities:
* [PERSON] Linguo 0.6144
* [PERSON] Lisa Simpson 1.4056

Unknown entities:
* Season 12 (1)
```
