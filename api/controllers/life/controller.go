package life

import (
	"bytes"
	"encoding/json"
	"io"

	"life_easy_api/api/middleware"
	"life_easy_api/api/schemas"
	"life_easy_api/api/services"
	"life_easy_api/pkg/db-pkg/models"

	"github.com/gofiber/fiber/v2"
)

type Controller struct {
	Service *services.Service
}

func bind(c *fiber.Ctx, input any) error {
	decoder := json.NewDecoder(bytes.NewReader(c.Body()))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(input); err != nil {
		return models.ErrInvalid
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return models.ErrInvalid
	}

	return nil
}

func (h *Controller) Register(c *fiber.Ctx) error {
	var input schemas.Register
	if err := bind(c, &input); err != nil {
		return err
	}

	result, err := h.Service.Register(c.UserContext(), input)
	if err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{"data": result})
}

func (h *Controller) Login(c *fiber.Ctx) error {
	var input schemas.Login
	if err := bind(c, &input); err != nil {
		return err
	}

	result, err := h.Service.Login(c.UserContext(), input)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *Controller) Workspace(c *fiber.Ctx) error {
	result, err := h.Service.Workspace(
		c.UserContext(),
		middleware.UserID(c),
	)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"data": result})
}

func (h *Controller) Save(c *fiber.Ctx) error {
	var input schemas.EntryInput
	if err := bind(c, &input); err != nil {
		return err
	}

	if err := h.Service.SaveEntry(
		c.UserContext(),
		middleware.UserID(c),
		c.Params("id"),
		input,
	); err != nil {
		return err
	}

	return h.Workspace(c)
}

func (h *Controller) Remove(c *fiber.Ctx) error {
	if err := h.Service.RemoveEntry(
		c.UserContext(),
		middleware.UserID(c),
		c.Params("id"),
	); err != nil {
		return err
	}

	return h.Workspace(c)
}

func (h *Controller) Toggle(c *fiber.Ctx) error {
	if err := h.Service.ToggleTask(
		c.UserContext(),
		middleware.UserID(c),
		c.Params("id"),
	); err != nil {
		return err
	}

	return h.Workspace(c)
}

func (h *Controller) Pay(c *fiber.Ctx) error {
	if err := h.Service.PayBill(
		c.UserContext(),
		middleware.UserID(c),
		c.Params("id"),
	); err != nil {
		return err
	}

	return h.Workspace(c)
}

func (h *Controller) CreateFamily(c *fiber.Ctx) error {
	var input schemas.CreateFamily
	if err := bind(c, &input); err != nil {
		return err
	}

	id, err := h.Service.CreateFamily(
		c.UserContext(),
		middleware.UserID(c),
		input,
	)
	if err != nil {
		return err
	}

	return c.Status(201).JSON(fiber.Map{
		"data": fiber.Map{"family_id": id},
	})
}

func (h *Controller) AddMember(c *fiber.Ctx) error {
	var input schemas.AddMember
	if err := bind(c, &input); err != nil {
		return err
	}

	if err := h.Service.AddMember(
		c.UserContext(),
		middleware.UserID(c),
		c.Params("id"),
		input,
	); err != nil {
		return err
	}

	return h.Workspace(c)
}

func (h *Controller) RemoveMember(c *fiber.Ctx) error {
	if err := h.Service.RemoveMember(
		c.UserContext(),
		middleware.UserID(c),
		c.Params("id"),
		c.Params("userID"),
	); err != nil {
		return err
	}

	return h.Workspace(c)
}
