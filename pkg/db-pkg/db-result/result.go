package dbresult

import "database/sql"

// RequireAffected คืน sql.ErrNoRows
// เมื่อไม่มีรายการที่ตรงกับเงื่อนไขและสิทธิ์ของผู้ใช้
func RequireAffected(result sql.Result, err error) error {
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return sql.ErrNoRows
	}

	return nil
}
