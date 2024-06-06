package plug

type Controller interface {
	Index(*Conn)  // index: show a list of object
	Edit(*Conn)   // edit: show edit form
	New(*Conn)    // new: show create object form
	Show(*Conn)   // show: show one object detail by id
	Create(*Conn) // create: save a new object
	Update(*Conn) // update: save update object
	Delete(*Conn) // delete: delete a object by id
}
