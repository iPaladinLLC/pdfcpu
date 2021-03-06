/*
Copyright 2018 The pdfcpu Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pdfcpu

import (
	"strings"
	"time"

	"github.com/iPaladinLLC/pdfcpu/pkg/log"
	"github.com/pkg/errors"
)

func textString(ctx *PDFContext, obj PDFObject) (string, error) {

	var s string
	var err error

	if indRef, ok := obj.(PDFIndirectRef); ok {
		obj, err = ctx.Dereference(indRef)
		if err != nil {
			return s, err
		}
	}

	obj, err = ctx.Dereference(obj)
	if err != nil {
		return s, err
	}

	switch obj := obj.(type) {

	case PDFStringLiteral:
		s, err = StringLiteralToString(obj.Value())
		if err != nil {
			return s, err
		}

	case PDFHexLiteral:
		s, err = HexLiteralToString(obj.Value())
		if err != nil {
			return s, err
		}

	default:
		return s, errors.Errorf("textString: corrupt -  %v\n", obj)
	}

	// Return a csv safe string.
	return strings.Replace(s, ";", ",", -1), nil
}

func writeInfoDict(ctx *PDFContext, dict *PDFDict) (err error) {

	for key, value := range dict.Dict {

		switch key {

		case "Title":
			log.Debug.Println("found Title")

		case "Author":
			log.Debug.Println("found Author")
			ctx.Author, err = textString(ctx, value)
			if err != nil {
				return err
			}

		case "Subject":
			log.Debug.Println("found Subject")

		case "Keywords":
			log.Debug.Println("found Keywords")

		case "Creator":
			log.Debug.Println("found Creator")
			ctx.Creator, err = textString(ctx, value)
			if err != nil {
				return err
			}

		case "Producer", "CreationDate", "ModDate":
			log.Debug.Printf("found %s", key)
			if indRef, ok := value.(PDFIndirectRef); ok {
				// Do not write indRef, will be modified by pdfcpu.
				ctx.Optimize.DuplicateInfoObjects[int(indRef.ObjectNumber)] = true
			}

		case "Trapped":
			log.Debug.Println("found Trapped")

		default:
			log.Debug.Printf("writeInfoDict: found out of spec entry %s %v\n", key, value)

		}
	}

	return nil
}

// Write the document info object for this PDF file.
// Add pdfcpu as Producer with proper creation date and mod date.
func writeDocumentInfoDict(ctx *PDFContext) error {

	// => 14.3.3 Document Information Dictionary

	// Optional:
	// Title                -
	// Author               -
	// Subject              -
	// Keywords             -
	// Creator              -
	// Producer		        modified by pdfcpu
	// CreationDate	        modified by pdfcpu
	// ModDate		        modified by pdfcpu
	// Trapped              -

	log.Debug.Printf("*** writeDocumentInfoDict begin: offset=%d ***\n", ctx.Write.Offset)

	// Document info object is optional.
	if ctx.Info == nil {
		log.Debug.Printf("writeDocumentInfoObject end: No info object present, offset=%d\n", ctx.Write.Offset)
		// Note: We could generate an info object from scratch in this scenario.
		return nil
	}

	log.Debug.Printf("writeDocumentInfoObject: %s\n", *ctx.Info)

	obj := *ctx.Info

	dict, err := ctx.DereferenceDict(obj)
	if err != nil || dict == nil {
		return err
	}

	// TODO Refactor - for stats only.
	err = writeInfoDict(ctx, dict)
	if err != nil {
		return err
	}

	// These are the modifications for the info dict of this PDF file:

	dateStringLiteral := DateStringLiteral(time.Now())

	dict.Update("CreationDate", dateStringLiteral)
	dict.Update("ModDate", dateStringLiteral)
	dict.Update("Producer", PDFStringLiteral(PDFCPULongVersion))

	_, _, err = writeDeepObject(ctx, obj)
	if err != nil {
		return err
	}

	log.Debug.Printf("*** writeDocumentInfoDict end: offset=%d ***\n", ctx.Write.Offset)

	return nil
}
