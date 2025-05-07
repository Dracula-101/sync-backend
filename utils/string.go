package utils

import "strconv"

func StringToInt(str string) (int, error) {
	if str == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func StringToInt64(str string) (int64, error) {
	if str == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func StringToUint(str string) (uint, error) {
	if str == "" {
		return 0, nil
	}
	i, err := strconv.ParseUint(str, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(i), nil
}

func StringToUint64(str string) (uint64, error) {
	if str == "" {
		return 0, nil
	}
	i, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}

func StringToBool(str string) (bool, error) {
	if str == "" {
		return false, nil
	}
	b, err := strconv.ParseBool(str)
	if err != nil {
		return false, err
	}
	return b, nil
}

func StringToFloat64(str string) (float64, error) {
	if str == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func StringToFloat32(str string) (float32, error) {
	if str == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(str, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

func StringToInt32(str string) (int32, error) {
	if str == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func StringToInt16(str string) (int16, error) {
	if str == "" {
		return 0, nil
	}
	i, err := strconv.ParseInt(str, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(i), nil
}
