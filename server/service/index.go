package service

import (
	"github.com/cryptogateway/backend-envoys/assets"
)

type IndexService struct {
	Context *assets.Context
}

func (i *IndexService) getPrice(base, quote string) (price, ratio interface{}, err error) {

	var (
		scales []float64
	)

	rows, err := i.Context.Db.Query("select price from trades where base_unit = $1 and quote_unit = $2 order by id desc limit 2", base, quote)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var (
			current float64
		)

		if err := rows.Scan(&current); err != nil {
			return nil, nil, err
		}

		scales = append(scales, current)
	}

	if len(scales) == 2 {
		ratio = ((scales[0] - scales[1]) / scales[1]) * 100
		price = scales[1]
	}

	return price, ratio, nil
}
