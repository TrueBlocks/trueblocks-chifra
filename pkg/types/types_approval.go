// Copyright 2016, 2024 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.
/*
 * Parts of this file were auto generated. Edit only those parts of
 * the code inside of 'EXISTING_CODE' tags.
 */

package types

// EXISTING_CODE
import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/TrueBlocks/trueblocks-chifra/v6/pkg/base"
)

// EXISTING_CODE

type Approval struct {
	Allowance    base.Wei       `json:"allowance"`
	BlockNumber  base.Blknum    `json:"blockNumber"`
	LastAppBlock base.Blknum    `json:"lastAppBlock"`
	LastAppLogID base.Lognum    `json:"lastAppLogID"`
	LastAppTs    base.Timestamp `json:"lastAppTs"`
	LastAppTxID  base.Txnum     `json:"lastAppTxID"`
	Owner        base.Address   `json:"owner"`
	OwnerName    string         `json:"ownerName,omitempty"`
	Spender      base.Address   `json:"spender"`
	SpenderName  string         `json:"spenderName,omitempty"`
	Timestamp    base.Timestamp `json:"timestamp"`
	Token        base.Address   `json:"token"`
	TokenName    string         `json:"tokenName,omitempty"`
	Calcs        *ApprovalCalcs `json:"calcs,omitempty"`
	// EXISTING_CODE
	// EXISTING_CODE
}

func (s Approval) String() string {
	bytes, _ := json.Marshal(s)
	return string(bytes)
}

func (s *Approval) Model(chain, format string, verbose bool, extraOpts map[string]any) Model {
	var order []string

	props := NewModelProps(chain, format, verbose, extraOpts)

	rawNames := []Labeler{
		NewLabeler(s.Owner, "owner"),
		NewLabeler(s.Spender, "spender"),
		NewLabeler(s.Token, "token"),
	}
	model := s.RawMap(props, &rawNames)
	for k, v := range s.CalcMap(props) {
		model[k] = v
	}

	// EXISTING_CODE
	order = []string{
		"blockNumber",
		"timestamp",
		"date",
		"owner",
		"spender",
		"token",
		"allowance",
		"lastAppBlock",
		"lastAppLogID",
		"lastAppTxID",
		"lastAppTs",
		"lastAppDate",
	}
	// EXISTING_CODE

	for _, item := range rawNames {
		key := item.name + "Name"
		if _, exists := model[key]; exists {
			order = append(order, key)
		}
	}
	order = reorderFields(order)

	return Model{
		Data:  model,
		Order: order,
	}
}

// RawMap returns a map containing only the raw/base fields for this Approval.
func (s *Approval) RawMap(p *ModelProps, needed *[]Labeler) map[string]any {
	model := map[string]any{
		// EXISTING_CODE
		"blockNumber":  s.BlockNumber,
		"timestamp":    s.Timestamp,
		"owner":        s.Owner,
		"spender":      s.Spender,
		"token":        s.Token,
		"allowance":    s.Allowance.String(),
		"lastAppBlock": s.LastAppBlock,
		"lastAppLogID": s.LastAppLogID,
		"lastAppTxID":  s.LastAppTxID,
		"lastAppTs":    s.LastAppTs,
		// EXISTING_CODE
	}

	// EXISTING_CODE
	if p.Verbose {
		return labelAddresses(p, model, needed)
	}
	// EXISTING_CODE

	return labelAddresses(p, model, needed)
}

// CalcMap returns a map containing the calculated/derived fields for this type.
func (s *Approval) CalcMap(p *ModelProps) map[string]any {
	_ = p // delint
	model := map[string]any{
		// EXISTING_CODE
		"date":        s.Date(),
		"lastAppDate": base.FormattedDate(s.LastAppTs),
		// EXISTING_CODE
	}

	// EXISTING_CODE
	// EXISTING_CODE

	return model
}

func (s *Approval) Date() string {
	return base.FormattedDate(s.Timestamp)
}

func (s *Approval) CacheLocations() (string, string, string) {
	paddedId := fmt.Sprintf("%s-%s-%s-%09d", s.Owner.Hex()[2:], s.Token.Hex()[2:], s.Spender.Hex()[2:], s.BlockNumber)
	parts := make([]string, 3)
	parts[0] = paddedId[:2]
	parts[1] = paddedId[2:4]
	parts[2] = paddedId[4:6]
	subFolder := strings.ToLower("Approval") + "s"
	directory := filepath.Join(subFolder, filepath.Join(parts...))
	return directory, paddedId, "bin"
}

func (s *Approval) MarshalCache(writer io.Writer) (err error) {
	// Allowance
	if err = base.WriteValue(writer, &s.Allowance); err != nil {
		return err
	}

	// BlockNumber
	if err = base.WriteValue(writer, s.BlockNumber); err != nil {
		return err
	}

	// LastAppBlock
	if err = base.WriteValue(writer, s.LastAppBlock); err != nil {
		return err
	}

	// LastAppLogID
	if err = base.WriteValue(writer, s.LastAppLogID); err != nil {
		return err
	}

	// LastAppTs
	if err = base.WriteValue(writer, s.LastAppTs); err != nil {
		return err
	}

	// LastAppTxID
	if err = base.WriteValue(writer, s.LastAppTxID); err != nil {
		return err
	}

	// Owner
	if err = base.WriteValue(writer, s.Owner); err != nil {
		return err
	}

	// Spender
	if err = base.WriteValue(writer, s.Spender); err != nil {
		return err
	}

	// Timestamp
	if err = base.WriteValue(writer, s.Timestamp); err != nil {
		return err
	}

	// Token
	if err = base.WriteValue(writer, s.Token); err != nil {
		return err
	}

	return nil
}

func (s *Approval) UnmarshalCache(fileVersion uint64, reader io.Reader) (err error) {
	// Check for compatibility and return cache.ErrIncompatibleVersion to invalidate this item (see #3638)
	// EXISTING_CODE
	// EXISTING_CODE

	// Allowance
	if err = base.ReadValue(reader, &s.Allowance, fileVersion); err != nil {
		return err
	}

	// BlockNumber
	if err = base.ReadValue(reader, &s.BlockNumber, fileVersion); err != nil {
		return err
	}

	// LastAppBlock
	if err = base.ReadValue(reader, &s.LastAppBlock, fileVersion); err != nil {
		return err
	}

	// LastAppLogID
	if err = base.ReadValue(reader, &s.LastAppLogID, fileVersion); err != nil {
		return err
	}

	// LastAppTs
	if err = base.ReadValue(reader, &s.LastAppTs, fileVersion); err != nil {
		return err
	}

	// LastAppTxID
	if err = base.ReadValue(reader, &s.LastAppTxID, fileVersion); err != nil {
		return err
	}

	// Owner
	if err = base.ReadValue(reader, &s.Owner, fileVersion); err != nil {
		return err
	}

	// Spender
	if err = base.ReadValue(reader, &s.Spender, fileVersion); err != nil {
		return err
	}

	// Timestamp
	if err = base.ReadValue(reader, &s.Timestamp, fileVersion); err != nil {
		return err
	}

	// Token
	if err = base.ReadValue(reader, &s.Token, fileVersion); err != nil {
		return err
	}

	s.FinishUnmarshal(fileVersion)

	return nil
}

// FinishUnmarshal is used by the cache. It may be unused depending on auto-code-gen
func (s *Approval) FinishUnmarshal(fileVersion uint64) {
	_ = fileVersion
	s.Calcs = nil
	// EXISTING_CODE
	// EXISTING_CODE
}

// ApprovalCalcs holds lazy-loaded calculated fields for Approval
type ApprovalCalcs struct {
	// EXISTING_CODE
	Date        string `json:"date"`
	LastAppDate string `json:"lastAppDate"`
	// EXISTING_CODE
}

func (s *Approval) EnsureCalcs(p *ModelProps, fieldFilter []string) error {
	_ = fieldFilter // delint
	if s.Calcs != nil {
		return nil
	}

	calcMap := s.CalcMap(p)
	if len(calcMap) == 0 {
		return nil
	}

	jsonBytes, err := json.Marshal(calcMap)
	if err != nil {
		return err
	}

	s.Calcs = &ApprovalCalcs{}
	return json.Unmarshal(jsonBytes, s.Calcs)
}

// EXISTING_CODE
// EXISTING_CODE
