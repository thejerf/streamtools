package streamtools

// A Tag is something that can be placed on data coming back from
// TaggedReaders indicating more information about them.
//
// In order to retain composability, the Tag interface is modelled on the
// errors package of more modern Go, with a generalized ability to wrap
// tags with other tags and a generalized ability to query into them with
// utility functions in this package.
//
// However, it is simply a fact that there is no way to specify readers to
// be infinitely composible. For instance, if you have a reader that is
// matching regular expressions and tagging them as being a match of
// something, but a composed TaggedReader sits on top of that and then
// further disassembles the resulting match and spits it back as the result
// of two separate calls, there simply is no correct way to propagate the
// tag from the regular expression matcher up the decorator stack any
// longer; the assertion the regular expression tag makes about what the
// match is is now fundamentally severed from the underlying
// data. TaggerReaders that compose other TaggerReaders are encouraged to
// carefully document exactly what they do to the tags passed through them.
type Tag interface {
	Unwrap() []Tag
}

// TagIs is an optional interface that tags can implement to indicate that
// they are the same as the passed-in comparison tag.
type TagIs interface {
	Is(t Tag) bool
}

var tagType = reflect.TypeOf((*Tag)(nil)).Elem()

func TagIs(tag Tag, target Tag) bool {
	if target == nil {
		return tag == target
	}

	isComparable := reflect.TypeOf(target).Comparable()

	if isComparable && tag == target {
		return true
	}

	if isable, ok := tag.(TagIs); ok && isable.Is(target) {
		return true
	}

	for _, subTag := tag.Unwrap() {
		if TagIs(subTag, target) {
			return true
		}
	}

	return false
}

func TagAs(tag Tag, target any) bool {
	if tag == nil {
		return false
	}

	if target == nil {
		panic("streamtools: target can not be nil")
	}

	val := reflect.ValueOf(target)
	typ := val.Type()
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		panic("streamtools: target must be a non-nil pointer")
	}
	targetType := typ.Elem()
	if targetType.Kind() != reflect.Interface && !targetType.Implements(tagType) {
		panic("streamtools: *target must be interface or implement Tag")
	}

	if reflect.TypeOf(tag).AssignableTo(targetType) {
		val.Elem().Set(reflect.ValueOf(tag))
		return true
	}

	if x, ok := tag.(interface{ As(any) bool }); ok && x.As(target) {
		return true
	}

	for _, subTag := range tag.Unwrap() {
		if As(subTag, target) {
			return true
		}
	}

	return false
}

// A TaggedReader is like an io.Reader, but it can return additional
// metadata about the bytes being returned, such as a tag indicating
// information the match in a regexp or something.
type TaggedReader interface {
	Read([]byte) (int, Tag, error)
}
