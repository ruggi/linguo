package main

import (
	"fmt"

	"github.com/federicoruggi/linguo"
)

func main() {
	engine := linguo.NewEngine()
	engine.InitNLP("./data", "en")

	result := engine.NLP.Workflow("Linguo was a grammar-correcting robot created by Lisa Simpson. It is from the eighteenth episode of Season 12.")

	for _, sentence := range result.Sentences {
		fmt.Printf("Sentence: \"%s\"\n", sentence.Body)
		fmt.Println("Entities:")
		for _, t := range sentence.Tokens {
			fmt.Printf("\t* [%s] %s (%s) %.4f%%\n", t.Pos, t.Base, t.Lemma, t.Prob)
		}
		fmt.Println()
	}

	fmt.Println("Entities:")
	for _, entity := range result.Entities {
		fmt.Printf("* [%s] %s %.4f\n", entity.Model, entity.Value, entity.Score)
	}
	fmt.Println()

	fmt.Println("Unknown entities:")
	for _, ue := range result.UnknownEntities {
		fmt.Printf("* %s (%d)\n", ue.Name, ue.Frequency)
	}
}
