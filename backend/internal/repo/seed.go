package repo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopuppy/internal/clock"
)

var (
	SeedDad    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	SeedMom    = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	SeedViewer = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	SeedFamily = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	SeedCream  = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	SeedBean   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

// Hash of Puppy123! bcrypt cost 10
const seedHash = "$2b$10$UYVmJB/bOFZ9Yu2dV4mQmu3Xy5H22Kkw94VO3LhQ8vkf0ce96A5.G"

func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	var n int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	now := clock.Now()
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `INSERT INTO users(id,email,password_hash,nickname,avatar_url,created_at,updated_at) VALUES
		($1,'dad@gopuppy.test',$4,'林爸爸','',$5,$5),
		($2,'mom@gopuppy.test',$4,'林妈妈','',$5,$5),
		($3,'viewer@gopuppy.test',$4,'寄养阿姨','',$5,$5)`,
		SeedDad, SeedMom, SeedViewer, seedHash, now)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO families(id,name,owner_id,created_at) VALUES ($1,'林家小院',$2,$3)`, SeedFamily, SeedDad, now); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO family_members(family_id,user_id,role,joined_at) VALUES
		($1,$2,'OWNER',$4),($1,$3,'CAREGIVER',$4),($1,$5,'VIEWER',$4)`,
		SeedFamily, SeedDad, SeedMom, now, SeedViewer); err != nil {
		return err
	}

	creamB := time.Date(2023, 3, 15, 0, 0, 0, 0, clock.Beijing)
	beanB := time.Date(2022, 8, 20, 0, 0, 0, 0, clock.Beijing)
	cmin, cmax := 4.2, 5.4
	bmin, bmax := 10.0, 13.5
	if _, err = tx.Exec(ctx, `INSERT INTO pets(id,family_id,name,species,breed,gender,birthday,avatar_key,neutered,chip_no,weight_min,weight_max,note,created_at) VALUES
		($1,$3,'奶油','CAT','英国短毛猫','FEMALE',$4,'',TRUE,'CHIP-CREAM-001',$6,$7,'家里的女王，喜欢窗台晒太阳',$10),
		($2,$3,'豆豆','DOG','柯基','MALE',$5,'',TRUE,'CHIP-BEAN-002',$8,$9,'短腿发动机，晚饭后必须散步',$10)`,
		SeedCream, SeedBean, SeedFamily, creamB, beanB, cmin, cmax, bmin, bmax, now); err != nil {
		return err
	}

	if _, err = tx.Exec(ctx, `INSERT INTO health_events(id,pet_id,category,title,description,occurred_at,clinic,severity,treated,amount_cents,created_by,created_at) VALUES
		($1,$3,'SURGERY','绝育手术','腹腔镜绝育，恢复良好', TIMESTAMPTZ '2024-03-12 10:00:00+08','萌宠医院','',TRUE,128000,$5,$7),
		($2,$3,'VACCINE','狂犬疫苗','年度加强针', TIMESTAMPTZ '2025-06-18 15:30:00+08','萌宠医院','',TRUE,18000,$5,$7),
		($8,$3,'DEWORM','体内驱虫','拜耳体内驱虫', TIMESTAMPTZ '2026-07-20 09:00:00+08','家庭','',TRUE,4500,$5,$7),
		($9,$3,'SYMPTOM','软便一天','晚上软便一次，次日恢复', TIMESTAMPTZ '2026-05-03 21:10:00+08','','MILD',FALSE,NULL,$6,$7),
		($10,$4,'VACCINE','狂犬疫苗','年度加强针', TIMESTAMPTZ '2025-09-02 11:00:00+08','萌宠医院','',TRUE,18000,$5,$7),
		($11,$4,'CHECKUP','年度体检','血常规+生化均正常', TIMESTAMPTZ '2026-01-16 14:00:00+08','萌宠医院','',TRUE,36000,$6,$7)`,
		uuid.New(), uuid.New(), SeedCream, SeedBean, SeedDad, SeedMom, now,
		uuid.New(), uuid.New(), uuid.New(), uuid.New()); err != nil {
		return err
	}

	nextVac := time.Date(2026, 6, 18, 0, 0, 0, 0, clock.Beijing)
	nextDew := time.Date(2026, 10, 20, 0, 0, 0, 0, clock.Beijing)
	if _, err = tx.Exec(ctx, `INSERT INTO reminder_rules(id,pet_id,kind,title,cycle_days,last_done_at,next_due_at,advance_days,channels,enabled,created_at) VALUES
		($1,$3,'VACCINE','奶油狂犬加强',365, TIMESTAMPTZ '2025-06-18 15:30:00+08',$5,3,ARRAY['EMAIL','WECOM_BOT'],TRUE,$7),
		($2,$3,'DEWORM','奶油体内驱虫',90, TIMESTAMPTZ '2026-07-20 09:00:00+08',$6,3,ARRAY['EMAIL','WEBHOOK'],TRUE,$7),
		($8,$4,'CHECKUP','豆豆年度体检',365, TIMESTAMPTZ '2026-01-16 14:00:00+08', DATE '2027-01-16',3,ARRAY['EMAIL'],TRUE,$7)`,
		uuid.New(), uuid.New(), SeedCream, SeedBean, nextVac, nextDew, now, uuid.New()); err != nil {
		return err
	}

	start := time.Date(now.Year(), now.Month(), 15, 10, 0, 0, 0, clock.Beijing).AddDate(0, -11, 0)
	for i := 0; i < 12; i++ {
		t := start.AddDate(0, i, 0)
		cw := 4.6 + float64(i%4)*0.12
		bw := 11.2 + float64(i%5)*0.18
		if _, err = tx.Exec(ctx, `INSERT INTO weight_records(id,pet_id,weight_kg,measured_at,note,created_by) VALUES
			($1,$3,$5,$7,'月度称重',$8),($2,$4,$6,$7,'月度称重',$8)`,
			uuid.New(), uuid.New(), SeedCream, SeedBean, cw, bw, t, SeedDad); err != nil {
			return err
		}
		food := int64(16800 + i*300)
		med := int64(0)
		if i%3 == 0 {
			med = 8900
		}
		toy := int64(3200)
		if _, err = tx.Exec(ctx, `INSERT INTO expenses(id,pet_id,category,amount_cents,spent_at,note,created_by) VALUES
			($1,$5,'FOOD',$7,$9,'猫粮',$11),
			($2,$5,'TOY',$8,$9,'逗猫棒',$11),
			($3,$6,'FOOD',$7+$8,$9,'狗粮',$11),
			($4,$6,'MEDICAL',$10,$9,'护理',$12)`,
			uuid.New(), uuid.New(), uuid.New(), uuid.New(), SeedCream, SeedBean, food, toy, t, med, SeedDad, SeedMom); err != nil {
			return err
		}
	}

	today := clock.Today()
	if _, err = tx.Exec(ctx, `INSERT INTO daily_checkins(id,pet_id,checkin_date,type,slot,done_by,done_at) VALUES
		($1,$2,$3,'FEED','MORNING',$4,$5)`,
		uuid.New(), SeedCream, today, SeedDad, now); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
