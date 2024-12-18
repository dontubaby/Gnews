package postgress

import (
	models "Skillfactory/36-GoNews/pkg/storage/models"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	_, err := New()
	if err != nil {
		t.Fatalf("Error of creating DB instance - %v", err)
	}
}

func TestNewMock(t *testing.T) {
	_, err := NewMock()
	if err != nil {
		t.Fatalf("Error of creating DB Mock instance - %v", err)
	}
}

func TestAddNews(t *testing.T) {
	db, err := NewMock()
	if err != nil {
		t.Fatalf("Error create DB instance - %v", err)
	}

	initDataSqlQuery := `DELETE FROM news WHERE published IN (1729584999,1729584991);`

	link1 := fmt.Sprintf("https://github.com/" + strconv.Itoa(rand.Intn(999999999999999999)))
	link2 := fmt.Sprintf("https://github.com/stretchr/testify" + strconv.Itoa(rand.Intn(999999999999999999)))

	news := []models.NewsFullDetailed{
		{
			ID:        1729584999,
			Title:     "Test Title",
			Content:   "Some test content here",
			Preview:   "Some test preview here",
			Published: 1729584999,
			Link:      link1,
		},
		{
			ID:        1729584991,
			Title:     "Test Title2",
			Content:   "Some test content here2",
			Preview:   "Some test preview here2",
			Published: 1729584991,
			Link:      link2,
		},
	}

	err = db.AddNews(news)
	if err != nil {
		t.Fatalf("Error adding NewsFullDetailed in database - %v", err)
	}
	//Очищаем базу от тестовой записи
	_, err = db.Db.Exec(context.Background(), initDataSqlQuery)
	if err != nil {
		t.Fatalf("Error of deleting data from DB - %v", err)
	}
}

// Тест проверяет что БД отдает данные записанные под тестовым ID
func TestGetDetailedNews(t *testing.T) {
	type testCase struct {
		name         string
		inputID      int
		expectedNews models.NewsFullDetailed
		expectedErr  error
	}

	initDataSqlQuery := `DELETE FROM news WHERE id=172958499991;`

	testCases := []testCase{
		{
			name:    "Valid ID",
			inputID: 172958499991,
			expectedNews: models.NewsFullDetailed{
				ID:      172958499991,
				Title:   "Test Title",
				Content: "Some test content here",

				Published: 172958499991,
				Link:      "https://go.dev/play/#172958499991",
			},
			expectedErr: nil,
		},
		{
			name:         "Invalid ID",
			inputID:      -1,
			expectedNews: models.NewsFullDetailed{},
			expectedErr:  errors.New("invalid news ID"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewMock()
			require.NoError(t, err)

			// Add data to the database if it's needed
			if tc.inputID > 0 {
				_, err = db.Db.Exec(context.Background(), `INSERT INTO news (id,title,content,published,link) VALUES ($1,$2,$3,$4,$5);`,
					tc.expectedNews.ID, tc.expectedNews.Title, tc.expectedNews.Content, tc.expectedNews.Published, tc.expectedNews.Link)
				require.NoError(t, err)
			}

			got, err := db.GetDetailedNews(tc.inputID)
			if tc.expectedErr == nil {
				require.NoError(t, err)
				require.Equal(t, tc.expectedNews, got)
			} else {
				require.EqualError(t, err, tc.expectedErr.Error())
			}
			//Очищаем базу от тестовой записи
			_, err = db.Db.Exec(context.Background(), initDataSqlQuery)
			if err != nil {
				t.Fatalf("Error of deleting data from DB - %v", err)
			}
		})
	}
}

func TestGetNewsList(t *testing.T) {
	type testCase struct {
		name          string
		inputCount    int
		expectedNews  []models.NewsFullDetailed
		expectedErr   error
		insertDataSQL string
	}
	e := errors.New("invalid count of news")
	initDataSqlQuery := `DELETE FROM news WHERE id IN (1600000000, 2100000000, 2200000000, 2300000000, 2400000000, 2500000000);`

	testCases := []testCase{
		{
			name:       "Valid Count",
			inputCount: 1,
			expectedNews: []models.NewsFullDetailed{
				{ID: 1600000000, Title: "Title 1", Preview: "Preview 1", Published: 1600000000, Link: "https://example.com/news/1"},
			},
			expectedErr: nil,
			insertDataSQL: ` INSERT INTO news (id, title, preview, published, link) VALUES 
			(1600000000, 'Title 1', 'Preview 1', 1600000000, 'https://example.com/news/1');`,
		},
		{
			name:       "Zero Count",
			inputCount: 0,
			expectedNews: []models.NewsFullDetailed{
				{ID: 2100000000, Title: "Title 1", Preview: "Preview 1", Published: 2100000000, Link: "https://example.com/news/21"},
				{ID: 2200000000, Title: "Title 2", Preview: "Preview 2", Published: 2200000000, Link: "https://example.com/news/22"},
				{ID: 2300000000, Title: "Title 3", Preview: "Preview 3", Published: 2300000000, Link: "https://example.com/news/23"},
				{ID: 2400000000, Title: "Title 4", Preview: "Preview 4", Published: 2400000000, Link: "https://example.com/news/24"},
				{ID: 2500000000, Title: "Title 5", Preview: "Preview 5", Published: 2500000000, Link: "https://example.com/news/25"},
			},
			expectedErr: e,
			insertDataSQL: ` INSERT INTO news (id, title, preview, published, link) VALUES 
			(2100000000, 'Title 1', 'Preview 1', 2100000000, 'https://example.com/news/21'), 
			(2200000000, 'Title 2', 'Preview 2', 2200000000, 'https://example.com/news/22'), 
			(2300000000, 'Title 3', 'Preview 3', 2300000000, 'https://example.com/news/23'), 
			(2400000000, 'Title 4', 'Preview 4', 2400000000, 'https://example.com/news/24'), 
			(2500000000, 'Title 5', 'Preview 5', 2500000000, 'https://example.com/news/25'); `,
		},
		{
			name:          "Negative Count",
			inputCount:    -1,
			expectedNews:  []models.NewsFullDetailed{},
			expectedErr:   e,
			insertDataSQL: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db, err := NewMock()
			require.NoError(t, err)

			// Insert data into the database if necessary
			if tc.insertDataSQL != "" {
				_, err = db.Db.Exec(context.Background(), tc.insertDataSQL)
				require.NoError(t, err)
			}

			got, err := db.GetNewsList(tc.inputCount)

			if tc.expectedErr == nil {
				//fmt.Println(news)
				require.NoError(t, err)
				require.Equal(t, tc.expectedNews, got)
			} else {
				require.EqualError(t, err, tc.expectedErr.Error())
			}
			//Очищаем базу от тестовой записи
			_, err = db.Db.Exec(context.Background(), initDataSqlQuery)
			if err != nil {
				t.Fatalf("Error of deleting data from DB - %v", err)
			}
		})

	}

}

func TestPrevieMaker(t *testing.T) {
	type args struct {
		detailNews string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "<100 symbols test",
			args: args{
				detailNews: "1234567890",
			},
			want: "12345...",
		},
		{
			name: ">100 symbols test",
			args: args{
				detailNews: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB",
			},
			want: "AAAAAAAAAAAAAAAAAAAAAAAAA...",
		},
		{
			name: "Empty string test",
			args: args{
				detailNews: "",
			},
			want: "...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrevieMaker(tt.args.detailNews); got != tt.want {
				t.Errorf("PrevieMaker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStorage_FilterNewsByContent(t *testing.T) {
	type fields struct {
		Db *pgxpool.Pool
	}

	db, err := NewMock()
	if err != nil {
		t.Fatalf("failed creating DB instance - %v", err)
	}

	tests := []struct {
		name             string
		fields           fields
		args             string
		insertDataSQL    string
		initDataSqlQuery string
		want             []models.NewsFullDetailed
		wantErr          bool
	}{
		{
			name: "Content filter test",
			fields: fields{
				Db: db.Db,
			},
			args: "test filter",
			insertDataSQL: `INSERT INTO news (id,title,content,preview,published,link) VALUES 
			(999999999999999999, 'title test','test filter content', 'Preview test', 999999999999999999, 'https://example.com/');`,
			initDataSqlQuery: `DELETE FROM news WHERE id IN (999999999999999999);`,
			want: []models.NewsFullDetailed{
				{ID: 999999999999999999, Title: "title test", Preview: "Preview test", Published: 999999999999999999,
					Link: "https://example.com/",
				},
			},
			wantErr: false,
		},

		{
			name: "Title filter test",
			fields: fields{
				Db: db.Db,
			},
			args: "test filter",
			insertDataSQL: `INSERT INTO news (id,title,content,preview,published,link) VALUES 
			(999999999999999999, 'test filter title',' content', 'Preview test', 999999999999999999, 'https://example.com/');`,
			initDataSqlQuery: `DELETE FROM news WHERE id IN (999999999999999999);`,
			want: []models.NewsFullDetailed{
				{ID: 999999999999999999, Title: "test filter title", Preview: "Preview test", Published: 999999999999999999,
					Link: "https://example.com/",
				},
			},
			wantErr: false,
		},

		{
			name: "Preview filter test",
			fields: fields{
				Db: db.Db,
			},
			args: "test filter",
			insertDataSQL: `INSERT INTO news (id,title,content,preview,published,link) VALUES 
			(999999999999999999, 'title',' content', 'test filter Preview test', 999999999999999999, 'https://example.com/');`,
			initDataSqlQuery: `DELETE FROM news WHERE id IN (999999999999999999);`,
			want: []models.NewsFullDetailed{
				{ID: 999999999999999999, Title: "title", Preview: "test filter Preview test", Published: 999999999999999999,
					Link: "https://example.com/",
				},
			},
			wantErr: false,
		},

		{
			name: "Empty filter test",
			fields: fields{
				Db: db.Db,
			},
			args: "",
			insertDataSQL: `INSERT INTO news (id,title,content,preview,published,link) VALUES 
			(999999999999999999, 'title',' content', 'test filter Preview test', 999999999999999999, 'https://example.com/1'),
			(999999999999999998, 'title',' content', 'test filter Preview test', 999999999999999998, 'https://example.com/2'),
			(999999999999999997, 'title',' content', 'test filter Preview test', 999999999999999997, 'https://example.com/3');`,
			initDataSqlQuery: `DELETE FROM news WHERE id IN (999999999999999999, 999999999999999998, 999999999999999997);`,
			want: []models.NewsFullDetailed{
				{ID: 999999999999999999, Title: "title", Preview: "test filter Preview test", Published: 999999999999999999,
					Link: "https://example.com/1",
				},
				{ID: 999999999999999998, Title: "title", Preview: "test filter Preview test", Published: 999999999999999998,
					Link: "https://example.com/2",
				},
				{ID: 999999999999999997, Title: "title", Preview: "test filter Preview test", Published: 999999999999999997,
					Link: "https://example.com/3",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.insertDataSQL != "" {
				_, err = db.Db.Exec(context.Background(), tt.insertDataSQL)
				require.NoError(t, err)
			}
			got, err := db.FilterNewsByContent(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.FilterNewsByContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.FilterNewsByContent() = %v, want %v", got, tt.want)
			}
			//Очищаем базу от тестовой записи
			_, err = db.Db.Exec(context.Background(), tt.initDataSqlQuery)
			if err != nil {
				t.Fatalf("Error of deleting data from DB - %v", err)
			}

		})
	}
}

func TestStorage_FilterNewsByPublished(t *testing.T) {
	type fields struct {
		Db *pgxpool.Pool
	}

	db, err := NewMock()
	if err != nil {
		t.Fatalf("failed creating DB instance - %v", err)
	}

	tests := []struct {
		name             string
		fields           fields
		args             int
		insertDataSQL    string
		initDataSqlQuery string
		want             []models.NewsFullDetailed
		wantErr          bool
	}{
		{
			name: "Valid filter",
			fields: fields{
				Db: db.Db,
			},
			args: 123456789,
			insertDataSQL: `INSERT INTO news (id,title,content,preview,published,link) VALUES 
			(999999999999999999, 'title test','test filter content', 'Preview test', 123456789, 'https://example.com/');`,
			initDataSqlQuery: `DELETE FROM news WHERE id IN (999999999999999999);`,
			want: []models.NewsFullDetailed{
				{ID: 999999999999999999, Title: "title test", Preview: "Preview test", Published: 123456789,
					Link: "https://example.com/",
				},
			},
			wantErr: false,
		},

		{
			name: "Invalid filter",
			fields: fields{
				Db: db.Db,
			},
			args: -1,
			insertDataSQL: `INSERT INTO news (id,title,content,preview,published,link) VALUES
			(999999999999999999, 'test filter title',' content', 'Preview test', 1234567890, 'https://example.com/');`,
			initDataSqlQuery: `DELETE FROM news WHERE id IN (999999999999999999);`,
			want:             []models.NewsFullDetailed{},

			wantErr: false,
		},

		{
			name: "Empty filter test",
			fields: fields{
				Db: db.Db,
			},
			args: 0,
			insertDataSQL: `INSERT INTO news (id,title,content,preview,published,link) VALUES
			(999999999999999999, 'title',' content', 'test filter Preview test', 999999999999999999, 'https://example.com/1'),
			(999999999999999998, 'title',' content', 'test filter Preview test', 999999999999999998, 'https://example.com/2'),
			(999999999999999997, 'title',' content', 'test filter Preview test', 999999999999999997, 'https://example.com/3');`,
			initDataSqlQuery: `DELETE FROM news WHERE id IN (999999999999999999, 999999999999999998, 999999999999999997);`,
			want:             []models.NewsFullDetailed{},
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.insertDataSQL != "" {
				_, err = db.Db.Exec(context.Background(), tt.insertDataSQL)
				require.NoError(t, err)
			}
			got, err := db.FilterNewsByPublished(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Storage.FilterNewsByPublished() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Storage.FilterNewsByPublished() = %v, want %v", got, tt.want)
			}
			//Очищаем базу от тестовой записи
			_, err = db.Db.Exec(context.Background(), tt.initDataSqlQuery)
			if err != nil {
				t.Fatalf("Error of deleting data from DB - %v", err)
			}

		})
	}
}
