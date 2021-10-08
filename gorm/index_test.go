package gorm

import (
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIndexes(t *testing.T) {
	type Profile struct {
		ID      uint
		Refer   uint   `gorm:"uniqueIndex"`
		Name    string `gorm:"index;size:252"`
		Content []byte `gorm:"index;size:252"`
		Age     sql.NullInt64
		Email   string `gorm:"size:252;unique_index"`
		Role    string `gorm:"size:252"` // set field size to 252
		//MemberNumber *string `gorm:"unique;not null;size:252"` // set member number to unique and not null not yet supported
		//MemberNumber *string `gorm:"not null;size:252"` // set member number not null
		Num      int    `gorm:"AUTO_INCREMENT"`      // set num to auto incrementable
		Address  string `gorm:"index:addr;size:252"` // create index with name `addr` for address
		IgnoreMe int    `gorm:"-"`                   // ignore this field
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.AutoMigrate(&Profile{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	//myNumber := "myNumber"

	err = DB.Create(&Profile{
		Refer:   2,
		Name:    "name",
		Content: []byte(`content`),
		Age:     sql.NullInt64{},
		Email:   "my@email.it",
		Role:    "my_role",
		//MemberNumber: &myNumber,
		Address:  "my_address",
		IgnoreMe: 55,
	}).Error
	require.NoError(t, err)
}

func TestCompositeIndex(t *testing.T) {
	type User struct {
		ID     uint
		Name   string `gorm:"index:idx_member;size:252"`
		Number string `gorm:"index:idx_member;size:252"`
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.AutoMigrate(&User{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	err = DB.Create(&User{
		Name:   "name",
		Number: "number",
	}).Error

	require.NoError(t, err)
}

func TestForeignKeys(t *testing.T) {
	type Profile struct {
		ID       uint
		Name     string `gorm:"index;size:256"`
		MemberID uint
	}

	type Member struct {
		ID         uint
		Refer      uint    `gorm:"uniqueIndex"`
		Name       string  `gorm:"index;size:256"`
		NameUnique string  `gorm:"index:idx_name,unique;size:256"`
		Profile    Profile `gorm:"FOREIGNKEY:MemberID;References:Refer"`
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.AutoMigrate(&Profile{}, &Member{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	member := Member{Refer: 1, Name: "foreign_key_constraints", Profile: Profile{Name: "my_profile"}}

	DB.Create(&member)

	var profile Profile
	if err := DB.First(&profile, "id = ?", member.Profile.ID).Error; err != nil {
		t.Fatalf("failed to find profile, got error: %v", err)
	} else if profile.MemberID != member.ID {
		t.Fatalf("member id is not equal: expects: %v, got: %v", member.ID, profile.MemberID)
	}

	member.Profile = Profile{}
	DB.Model(&member).Update("Refer", 100)

	var profile2 Profile
	if err := DB.First(&profile2, "id = ?", profile.ID).Error; err != nil {
		t.Fatalf("failed to find profile, got error: %v", err)
	} else if profile2.MemberID != 100 {
		t.Fatalf("member id is not equal: expects: %v, got: %v", 100, profile2.MemberID)
	}

	if r := DB.Delete(&member); r.Error != nil || r.RowsAffected != 1 {
		t.Fatalf("Should delete member, got error: %v, affected: %v", r.Error, r.RowsAffected)
	}

	var result Member
	if err := DB.First(&result, member.ID).Error; err == nil {
		t.Fatalf("Should not find deleted member")
	}

	if err := DB.First(&profile2, profile.ID).Error; err == nil {
		t.Fatalf("Should not find deleted profile")
	}
}
