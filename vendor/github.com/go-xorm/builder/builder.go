// Copyright 2016 The Xorm Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package builder

type optype byte

const (
	condType   optype = iota // only conditions
	selectType               // select
	insertType               // insert
	updateType               // update
	deleteType               // delete
	unionType                // union
)

type join struct {
	joinType  string
	joinTable string
	joinCond  Cond
}

type union struct {
	unionType string
	builder   *Builder
}

// Builder describes a SQL statement
type Builder struct {
	optype
	tableName string
	subQuery  *Builder
	cond      Cond
	selects   []string
	joins     []join
	unions    []union
	inserts   Eq
	updates   []Eq
	orderBy   string
	groupBy   string
	having    string
}

// Where sets where SQL
func (b *Builder) Where(cond Cond) *Builder {
	b.cond = b.cond.And(cond)
	return b
}

// From sets the name of table or the sub query's alias and itself
func (b *Builder) From(tableName string, subQuery ...*Builder) *Builder {
	b.tableName = tableName

	if len(subQuery) > 0 {
		b.subQuery = subQuery[0]
	}

	return b
}

// TableName returns the table name
func (b *Builder) TableName() string {
	return b.tableName
}

// Into sets insert table name
func (b *Builder) Into(tableName string) *Builder {
	b.tableName = tableName
	return b
}

// Join sets join table and conditions
func (b *Builder) Join(joinType, joinTable string, joinCond interface{}) *Builder {
	switch joinCond.(type) {
	case Cond:
		b.joins = append(b.joins, join{joinType, joinTable, joinCond.(Cond)})
	case string:
		b.joins = append(b.joins, join{joinType, joinTable, Expr(joinCond.(string))})
	}

	return b
}

// Union sets union conditions
func (b *Builder) Union(unionTp string, unionCond *Builder) *Builder {
	var builder *Builder
	if b.optype != unionType {
		builder = &Builder{cond: NewCond()}
		builder.optype = unionType

		currentUnions := b.unions
		// erase sub unions (actually append to new Builder.unions)
		b.unions = nil

		builder.unions = append(append(builder.unions, union{"", b}), currentUnions...)
	} else {
		builder = b
	}

	if unionCond != nil {
		builder.unions = append(builder.unions, union{unionTp, unionCond})
	}

	return builder
}

// InnerJoin sets inner join
func (b *Builder) InnerJoin(joinTable string, joinCond interface{}) *Builder {
	return b.Join("INNER", joinTable, joinCond)
}

// LeftJoin sets left join SQL
func (b *Builder) LeftJoin(joinTable string, joinCond interface{}) *Builder {
	return b.Join("LEFT", joinTable, joinCond)
}

// RightJoin sets right join SQL
func (b *Builder) RightJoin(joinTable string, joinCond interface{}) *Builder {
	return b.Join("RIGHT", joinTable, joinCond)
}

// CrossJoin sets cross join SQL
func (b *Builder) CrossJoin(joinTable string, joinCond interface{}) *Builder {
	return b.Join("CROSS", joinTable, joinCond)
}

// FullJoin sets full join SQL
func (b *Builder) FullJoin(joinTable string, joinCond interface{}) *Builder {
	return b.Join("FULL", joinTable, joinCond)
}

// Select sets select SQL
func (b *Builder) Select(cols ...string) *Builder {
	b.selects = cols
	b.optype = selectType
	return b
}

// And sets AND condition
func (b *Builder) And(cond Cond) *Builder {
	b.cond = And(b.cond, cond)
	return b
}

// Or sets OR condition
func (b *Builder) Or(cond Cond) *Builder {
	b.cond = Or(b.cond, cond)
	return b
}

// Insert sets insert SQL
func (b *Builder) Insert(eq Eq) *Builder {
	b.inserts = eq
	b.optype = insertType
	return b
}

// Update sets update SQL
func (b *Builder) Update(updates ...Eq) *Builder {
	b.updates = make([]Eq, 0, len(updates))
	for _, update := range updates {
		if update.IsValid() {
			b.updates = append(b.updates, update)
		}
	}
	b.optype = updateType
	return b
}

// Delete sets delete SQL
func (b *Builder) Delete(conds ...Cond) *Builder {
	b.cond = b.cond.And(conds...)
	b.optype = deleteType
	return b
}

// WriteTo implements Writer interface
func (b *Builder) WriteTo(w Writer) error {
	switch b.optype {
	case condType:
		return b.cond.WriteTo(w)
	case selectType:
		return b.selectWriteTo(w)
	case insertType:
		return b.insertWriteTo(w)
	case updateType:
		return b.updateWriteTo(w)
	case deleteType:
		return b.deleteWriteTo(w)
	case unionType:
		return b.unionWriteTo(w)
	}

	return ErrNotSupportType
}

// ToSQL convert a builder to SQL and args
func (b *Builder) ToSQL() (string, []interface{}, error) {
	w := NewWriter()
	if err := b.WriteTo(w); err != nil {
		return "", nil, err
	}

	return w.writer.String(), w.args, nil
}

// ToBindedSQL
func (b *Builder) ToBindedSQL() (string, error) {
	w := NewWriter()
	if err := b.WriteTo(w); err != nil {
		return "", err
	}

	return ConvertToBindedSQL(w.writer.String(), w.args)
}
