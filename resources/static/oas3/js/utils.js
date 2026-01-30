class Utils {
    static oas3() {
        const key = 'oas3'
        const quires = new URLSearchParams(window.location.search)
        const name = quires.get(key)
        if (name) {
            return name
        }

        return sessionStorage.getItem(key)
    }

}
