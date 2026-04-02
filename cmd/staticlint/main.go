// Package main реализует multichecker для статического анализа Go-кода.
// Он объединяет:
//   - стандартные анализаторы из golang.org/x/tools/go/analysis/passes;
//   - все анализаторы класса SA из staticcheck.io;
//   - один анализатор класса ST из staticcheck.io;
//   - два публичных анализатора: ineffassign и errcheck;
//   - собственный анализатор, запрещающий прямой вызов os.Exit в функции main пакета main.
//
// Запуск:
//
//	$ go build -o multichecker main.go
//	$ ./multichecker ./...
//
// Флаги можно передавать как обычно, например:
//
//	$ ./multichecker -printf.funcs=MyPrintf ./...
package main

import (
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck/st1001"
	"honnef.co/go/tools/stylecheck/st1013"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/kisielk/errcheck/errcheck"
	"github.com/nk87rus/go-musthave-shortener/cmd/staticlint/noexit"
)

// main – точка входа в multichecker.
// Здесь собирается список всех анализаторов, которые будут запущены.
// Затем управление передаётся функции multichecker.Main.
func main() {
	// Собираем все анализаторы
	var analyzers []*analysis.Analyzer

	// 1. Стандартные анализаторы из golang.org/x/tools/go/analysis/passes.
	//    Каждый из них выполняет определённую проверку, рекомендованную командой Go.
	analyzers = append(analyzers,
		assign.Analyzer,           // находит неиспользуемые присваивания
		atomic.Analyzer,           // проверяет распространённые ошибки при использовании пакета sync/atomic
		bools.Analyzer,            // ищет ошибки в булевых выражениях (например, всегда истинные условия)
		composite.Analyzer,        // предупреждает о неименованных литералах составных типов
		copylock.Analyzer,         // обнаруживает копирование блокировок (mutex) по значению
		defers.Analyzer,           // проверяет распространённые ошибки в отложенных вызовах (defer)
		directive.Analyzer,        // проверяет корректность директив //go:...
		errorsas.Analyzer,         // проверяет, что второй аргумент errors.As является указателем на тип, реализующий error
		framepointer.Analyzer,     // проверяет ассемблерный код на сохранение указателя фрейма
		httpresponse.Analyzer,     // ищет ошибки при работе с HTTP-ответами (например, закрытие тела до чтения)
		ifaceassert.Analyzer,      // находит невозможные утверждения типа (type assertion) для интерфейсов
		loopclosure.Analyzer,      // обнаруживает захват переменных цикла в замыканиях
		lostcancel.Analyzer,       // находит случаи, когда функция отмены контекста не вызывается
		nilfunc.Analyzer,          // ищет бесполезные сравнения функций с nil
		printf.Analyzer,           // проверяет соответствие форматной строки и аргументов в функциях семейства Printf
		shadow.Analyzer,           // предупреждает о затенении переменных (переопределении во вложенных областях видимости)
		sigchanyzer.Analyzer,      // обнаруживает неправильное использование сигнальных каналов в signal.Notify
		slog.Analyzer,             // проверяет корректность пар "ключ-значение" в вызовах log/slog
		stdmethods.Analyzer,       // проверяет сигнатуры методов, похожих на известные интерфейсы (например, String() string)
		stringintconv.Analyzer,    // предупреждает о преобразованиях целых чисел в строку (string(123) → "\u007b")
		structtag.Analyzer,        // проверяет корректность тегов полей структур
		testinggoroutine.Analyzer, // находит вызовы Fatal из горутин тестов
		tests.Analyzer,            // проверяет распространённые ошибки в тестах и примерах
		unmarshal.Analyzer,        // проверяет, что в функции unmarshal передаются указатели или интерфейсы
		unreachable.Analyzer,      // находит недостижимый код
		unsafeptr.Analyzer,        // проверяет небезопасные преобразования uintptr в unsafe.Pointer
		unusedresult.Analyzer,     // находит неиспользуемые результаты вызовов некоторых функций (например, fmt.Errorf)
	)

	// 2. Все анализаторы класса SA из staticcheck.io.
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// 3. Анализаторы класса ST
	analyzers = append(analyzers,
		st1001.Analyzer, // точечный импорт не рекомендуется
		st1013.Analyzer, // cледует использовать константы для кодов ошибок HTTP, а не магические числа
	)

	// 4. Два публичных анализатора
	analyzers = append(analyzers, ineffassign.Analyzer) // поиск неиспользуемых присваиваний
	analyzers = append(analyzers, errcheck.Analyzer)    // проверка обработки ошибок

	// 5. noExit анализатор
	analyzers = append(analyzers, noexit.Analyzer)

	// Запуск multichecker.
	// Функция multichecker.Main принимает список анализаторов и инициирует
	// их выполнение на пакетах, указанных в аргументах командной строки.
	multichecker.Main(analyzers...)
}
