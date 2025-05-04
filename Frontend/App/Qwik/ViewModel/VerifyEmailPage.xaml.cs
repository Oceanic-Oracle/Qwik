using Qwik.Model;
using Qwik.Model.Dto;

namespace Qwik;

public partial class VerifyEmailPage : ContentPage
{
    private readonly string Login;
    private readonly string Email;
	private readonly string Password;
    private readonly string SessionCode;

    public VerifyEmailPage(string login, string email, string password, string sessionCode)
    {
        this.Login = login;
        this.Email = email;
        this.Password = password;
        this.SessionCode = sessionCode;
        InitializeComponent();
    }

    private async void ConfirmRegClicked(object sender, EventArgs e)
    {
        try
        {
            string verifyCode = VerifyCode.Text;

            if (string.IsNullOrEmpty(verifyCode))
            {
                await DisplayAlert("Ошибка", "Введите код подтверждения", "ОК");
                return;
            }

            ConfirmRegReq req = new ConfirmRegReq
            {
                sessionCode = SessionCode,
                login = Login,
                password = Password,
                email = Email,
                verifyCode = verifyCode
            };

            ConfirmRegRes res = await Api.SendPostRequest<ConfirmRegReq, ConfirmRegRes>(req, Api.ConfirmRegEndpoint);

            if (string.IsNullOrEmpty(res.id))
            {
                await DisplayAlert("Ошибка", "Неизветная ошибка", "ОК");
                return;
            }

            await Task.WhenAll(
                Shell.Current.Navigation.PopAsync(),
                Shell.Current.Navigation.PopAsync()
            );
        }
        catch (Exception ex)
        {
            await DisplayAlert("Ошибка", $"{ex}", "ОК");
            return;
        }
    }
}