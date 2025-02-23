package path

// ファイル、ディレクトリのパス文字列を扱うためのパッケージ

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// パス型
type Path string
type Entries []Path
type Ext string

// パスを作成
func NewPath(p string) Path {
	return Path(p)
}

// カレントディレクトリのパスを取得
func GetCurDir() Path {
	p, _ := os.Getwd()
	return NewPath(p)
}

// パスを文字列に変換
func (p Path) String() string {
	return string(p)
}

// 拡張子を文字列に変換
func (e Ext) String() string {
	return string(e)
}

// 拡張子を小文字に変換
func (e Ext) Lower() Ext {
	return Ext(strings.ToLower(string(e)))
}

// 拡張子を大文字に変換
func (e Ext) Upper() Ext {
	return Ext(strings.ToUpper(string(e)))
}

// パスの結合
func Join(element ...Path) Path {
	elements := make([]string, len(element))
	for i, e := range element {
		elements[i] = string(e)
	}
	return Path(filepath.Join(elements...))
}

// パスの結合
func (p *Path) Append(element ...Path) {
	*p = Join(append([]Path{*p}, element...)...)
}

// 最後の要素を取得
func (p Path) Base() Path {
	return Path(filepath.Base(string(p)))
}

// Path が存在するか判定
func (p Path) IsExist() bool {
	_, err := os.Stat(string(p))
	return err == nil
}

// Path がディレクトリか判定、存在しない場合は false
func (p Path) IsDir() bool {
	fi, err := os.Stat(string(p))
	if err != nil {
		return false
	}
	return fi.IsDir()
}

// Path がファイルか判定、存在しない場合は false
func (p Path) IsFile() bool {
	fi, err := os.Stat(string(p))
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

// 絶対パスを取得
func (p Path) Abs() (Path, error) {
	abs, err := filepath.Abs(string(p))
	if err != nil {
		return "", err
	}
	return Path(abs), nil
}

// 絶対パスに変更
func (p *Path) ChangeAbs() error {
	abs, err := p.Abs()
	if err != nil {
		return err
	}
	*p = abs
	return nil
}

// ディレクトリを作成
func (p Path) CreDir() error {
	if p.IsDir() {
		return nil
	}
	return os.MkdirAll(string(p), 0777)
}

// ディレクトリを削除
func (p Path) DelDir() error {
	if !p.IsDir() {
		return nil
	}
	return os.RemoveAll(string(p))
}

// ファイルを作成
func (p Path) CreFile() (*os.File, error) {
	if p.IsFile() {
		// 既にファイルが存在する場合はエラー
		return nil, os.ErrExist
	}
	// ファイルが存在しない場合は作成
	return os.Create(string(p))
}

// ファイルを削除
func (p Path) DelFile() error {
	if !p.IsFile() {
		return nil
	}
	return os.Remove(string(p))
}

// ファイルを開く
func (p Path) FileOpen() (*os.File, error) {
	// ファイルでない場合はエラー
	if !p.IsFile() {
		return nil, os.ErrNotExist
	}
	// ファイルを開く
	return os.Open(string(p))
}

// ディレクトリ名を取得
func (p Path) DirName() Path {
	return Path(filepath.Dir(string(p)))
}

// ファイル名を取得、拡張子を含む
func (p Path) FileName() Path {
	return Path(filepath.Base(string(p)))
}

// ファイル名を取得、拡張子を除く
func (p Path) FileNameWithoutExt() Path {
	if p.Ext() == "" {
		return p.Base()
	}
	return p.Base()[:len(p.Base())-len(p.Ext())]
}

// ファイル名を変更、拡張子は変更しない
func (p *Path) ChangeFileName(name Path) {
	*p = Join(p.DirName(), Path(name.String()+p.Ext().String()))
}

// ファイル名の後ろに文字列を追加、拡張子は変更しない
func (p *Path) AddPrefix(name string) {
	*p = Join(p.DirName(), Path(name+p.FileNameWithoutExt().String()+p.Ext().String()))
}

// ファイル名の前に文字列を追加、拡張子は変更しない
func (p *Path) AddSuffix(name string) {
	*p = Join(p.DirName(), Path(p.FileNameWithoutExt().String()+name+p.Ext().String()))
}

// ファイル名は変更せず、ディレクトリ名を変更
func (p *Path) ChangeDirName(dir Path) {
	*p = Join(dir, p.FileName())
}

// 拡張子の付与
func (p *Path) AddExt(ext Ext) {
	*p += Path(ext.String())
}

// 拡張子を取得
func (p Path) Ext() Ext {
	return Ext(filepath.Ext(string(p)))
}

// 拡張子を変更
func (p *Path) ChangeExt(ext Ext) {
	if ext == "" {
		// 新拡張子が空の場合は、現拡張子を削除
		*p = p.FileNameWithoutExt()
		return
	}
	if p.Ext() == "" {
		// 現拡張子がない場合は付与
		p.AddExt(ext)
	} else {
		*p = NewPath(p.FileNameWithoutExt().String() + ext.String())
	}
}

// 拡張子を小文字に変換
func (p *Path) LowerExt() {
	ext := p.Ext().Lower()
	p.ChangeExt(ext)
}

// 拡張子を大文字に変換
func (p *Path) UpperExt() {
	ext := p.Ext().Upper()
	p.ChangeExt(ext)
}

// ディレクトリ内のファイル、ディレクトリを取得
func (p Path) Entries() (Entries, error) {
	// ディレクトリでない場合はエラー
	if !p.IsDir() {
		return Entries{}, os.ErrNotExist
	}

	// ディレクトリを開く
	dir, err := os.Open(string(p))
	if err != nil {
		return Entries{}, err
	}
	defer dir.Close()

	// ディレクトリ内のファイル、ディレクトリを取得
	names, err := dir.Readdirnames(-1)
	if err != nil {
		return Entries{}, err
	}

	// パスを作成
	entries := make(Entries, len(names))
	for i, name := range names {
		entries[i] = Join(p, NewPath(name))
	}
	return entries, nil
}

// ディレクトリ内のファイル、ディレクトリを取得
func Grab(p Path) (Entries, error) {
	return p.Entries()
}

// Entries から抽出する一般処理
func (e Entries) Filter(f func(Path) bool) Entries {
	entries := Entries{}
	for _, entry := range e {
		if f(entry) {
			entries = append(entries, entry)
		}
	}
	return entries
}

// Entries から存在するもののみ抽出
func (e Entries) ExtractExist() Entries {
	return e.Filter(func(p Path) bool {
		return p.IsExist()
	})
}

// Entries からディレクトリのみ抽出、存在しないものは除外
func (e Entries) ExtractDirs() Entries {
	return e.Filter(func(p Path) bool {
		return p.IsDir()
	})
}

// Entries からファイルのみ抽出、存在しないものは除外
func (e Entries) ExtractFiles() Entries {
	return e.Filter(func(p Path) bool {
		return p.IsFile()
	})
}

// Entries から指定の拡張子のファイルのみ抽出
func (e Entries) ExtractExt(exts ...Ext) Entries {
	return e.Filter(func(p Path) bool {
		for _, ext := range exts {
			if p.Ext() == ext {
				return true
			}
		}
		return false
	})
}

// Entries を []string に変換
func (e Entries) ToString() []string {
	result := make([]string, len(e))
	for i, entry := range e {
		result[i] = entry.String()
	}
	return result
}

// Entries をすべて絶対パスに変換
func (e Entries) ToAbs() (Entries, error) {
	return e.ForEachWithError(func(p Path) (Path, error) {
		return p.Abs()
	})
}

// Entries からファイル名のみ抽出
func (e Entries) ToBase() Entries {
	return e.ForEach(func(p Path) Path {
		return p.Base()
	})
}

// Entries から拡張子のみ抽出、重複を除外
func (e Entries) ToExt() []Ext {
	// map で重複を除外
	extsMap := map[Ext]struct{}{}
	for _, entry := range e {
		extsMap[entry.Ext()] = struct{}{}
	}
	// ソートして返す
	result := make([]Ext, 0, len(extsMap))
	for ext := range extsMap {
		result = append(result, ext)
	}
	sort.Slice(result, func(i, j int) bool {
		return string(result[i]) < string(result[j])
	})
	return result
}

// Entries 全てに共通の処理を適用して返す。
func (e Entries) ForEach(proc func(Path) Path) Entries {
	neu := make(Entries, len(e))
	for i := 0; i < len(e); i++ {
		neu[i] = proc(e[i])
	}
	return neu
}

// Entries 全てに共通の処理を適用して返す。エラー対応版。
func (e Entries) ForEachWithError(proc func(Path) (Path, error)) (Entries, error) {
	neu := make(Entries, len(e))
	for i := 0; i < len(e); i++ {
		p, err := proc(e[i])
		if err != nil {
			return nil, err
		}
		neu[i] = p
	}
	return neu, nil
}

// Entries の全ての要素がファイルであると仮定し、各ファイルのファイル名に対して処理を適用する関数
func (e Entries) ForEachFileName(proc func(Path) Path) Entries {
	return e.ForEach(func(p Path) Path {
		dir := p.DirName()
		base := p.FileName()
		return Join(dir, proc(base))
	})
}

// PrependSequentialNumbers は、
// Entries の全てのファイル名の先頭に連番を付与して更新する関数です。
// ファイル数に応じて連番の桁数を自動設定します。
func (e Entries) PrependSequentialNumbers() Entries {
	digits := len(strconv.Itoa(len(e)))
	counter := 0
	neu := e.ForEachFileName(func(name Path) Path {
		counter++
		return NewPath(fmt.Sprintf("%0*d_%s", digits, counter, name))
	})
	return neu
}
